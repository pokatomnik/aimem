# Lessons

Corrections the user has made to this skill's output — findings it reported and
text it wrote. Read before every run.

Rules for this file:

- Entries are appended, never edited in place. A reversed decision gets a new
  entry and the old one is marked `(superseded by #N)`.
- Only real corrections by the user belong here. No speculation, no notes about
  what seemed odd.
- Each entry states what was done, what the user said, why it was wrong, the
  rule, and a check that proves the rule.
- A lesson about the audit itself (a wrong claim in `SKILL.md`) also fixes that
  section; the entry keeps the reasoning.

Nothing recorded yet.

---

## #1 — A fake error must wrap the sentinel, not copy its text

**What I did.** Wrote `TestRemoveFactByIDMissingFact` in
`internal/repositories/kv/getfact_test.go` with a fake `Delete` result whose
`Err` was `errors.New(`xvec: DELETE: db: primary key not found: "absent"`)` —
the message copied from a real run, but a fresh error with no relationship to
`xvec.ErrNotFound`.

**What the user said.** «Сделал как ты говоришь, все равно та же проблема» —
after correctly rewriting `RemoveFactByID` to use `errors.Is(v.Err, xvec.ErrNotFound)`.
The code was right; the test was wrong.

**Why it was wrong.** `errors.Is` compares identity, or calls an `Is` method —
never the message string. The real xvec error wraps the sentinel (verified:
`errors.Is(item.Err, xvec.ErrNotFound) == true`), so a text-only fake can never
match, and the test would have failed against a correct implementation. Worse,
it would also have passed for any implementation that inspected the message.

**Rule.** A fake error that the code compares with `errors.Is`/`errors.As` must
wrap the same sentinel the real dependency wraps (`fmt.Errorf("...: %w",
ErrNotFound)`). Reproduction must copy the _shape_ of the real error, not its
text.

**Check.** Revert the production file via `go test -overlay` and confirm the test
fails for the reason it describes; then restore and confirm it passes.

---

## #2 — A fake must copy the whole return shape, not just the part I remembered

**What I did.** Made the fake `Delete` return `nil` as its error and put the
not-found cause only into the per-key `WriteResult.Err`. The regression test then
went green against an implementation that still leaked xvec's batch error.

**What the user said.** «Я добавил, проверяй и пофикси тест» — after the earlier
round already went green while real `DELETE` on a missing fact still answered
`500 INTERNAL_ERROR` with `xvec: batch write: 1 document operations failed`.

**Why it was wrong.** Real `xvec.Delete` returns **both** a non-empty result list
**and** a batch error wrapping `xvec.ErrNotFound`, with the same cause repeated
per key. A fake that models only one of the two channels makes the test pass for
any implementation that reads that one channel — the other path stays unguarded.

**Rule.** When faking a dependency, reproduce every channel the real call returns
(values _and_ error, batch _and_ per-item), and confirm the fake against a real
run before trusting the test. A green test over a fake that is simpler than
reality proves nothing.

**Check.** `go test -overlay` against the pre-fix file must fail; and an
end-to-end run against the real binary must agree with the unit test. A unit test
and a live check disagreeing means the fake is wrong, not the live check.

---

## #3 — Report against the file, not against a diff hunk; and docs work is not a code review

**What I did.** In `internal/controllers/web/handler.go` I reported `mcpPrefix`
as declared-but-unused, because the `git diff` hunk showed the `const` block
adding it while the `switch` case I read used the literal `"/mcp"`. In the same
report I added three unsolicited code findings: a duplicated `net/http` import
alias in `internal/adapters/servers/http/http.go`, a deleted explanatory comment,
and a nil-`mcpHandler` hazard.

**What the user said.** «по коду - ОК, так и надо, константу я использовал,
остальное трогать не нужно» — and then asked only for the docs to be updated.

**Why it was wrong.** Two separate mechanisms. First, a diff hunk is a snapshot
of one edit against one base: it does not show what the file says now, and it
cannot show whether a symbol is used elsewhere in the file at all. Reading the
hunk and concluding "unused" skipped the one check that settles it — a `grep` of
the current file. Second, the user writes the code by hand and asks for text
work; piling on style findings about code they just wrote turns a docs task into
an unrequested review and buries the findings they did want.

**Rule.** Before reporting any code finding, confirm it against the current file
with `read_file`/`grep`, never against a diff hunk alone. When the user asks for
docs, report only what the docs get wrong; keep code observations to a short
"judgment calls" list, and drop them entirely if the user says the code is fine.

**Check.** `grep -c mcpPrefix internal/controllers/web/handler.go` returns 2 —
declaration and use. For any "unused symbol" claim, the same grep must come back
with a count of 1 (declaration only) before the claim is written down.

---

## #4 — A swallowed parse error is a logging decision, not a crash

**What I did.** Reported `internal/repositories/users/users.go:36-39`
(`if err != nil { return result }`) as bug B2: a malformed `AIMEM_USERS_JSON`
degrades to an empty user map, which surfaces as `ErrNoUsersConfigured`, and I
asked for a startup error before documenting anything.

**What the user said.** «Ошибка парсинга и должна проглатываться, здесь будет
решение на уровне логгирования, а не на уровне "упасть если JSON не парсится"».

**Why it was wrong.** I read "returns an empty map" as the end of the path and
assumed the defect was that nothing stopped the process. The path continues one
layer up: with Qdrant the empty map hits `len(configured) == 0` and answers
`ErrNoUsersConfigured`, and `scaffold` refuses to start on that error. So the
deployment is never served with a broken user list; what is missing is the cause
in the log, which is a separate concern the user intends to solve on its own.
Turning `New` into a second error channel would push that decision into the
repository.

**Rule.** Before calling a swallowed error a bug, follow the degraded value to
its caller. When the degraded value still stops the process, the finding is at
most a judgment call about the message, and it does not block documentation.

**Check.** `grep -n "ErrNoUsersConfigured" internal/repositories/users/users.go`
— the empty map reaches a branch whose error stops `scaffold`, so no request is
served against a half-parsed list.

---

## #5 — Duplicate tokens are accepted on purpose

**What I did.** Reported `users.go:44` (`result[token] = ...`, last wins) as bug
B3: a token listed under two users silently authenticates as the later one, and
the previous file-based lookup was first-wins, so the semantics changed silently.
I wanted duplicates rejected at parse time.

**What the user said.** «Дубликат токенов это ок, там не будет тысячи
пользователей, а если будет то решение добавлять пользователей через env уже не
будет подходить. Сейчас это избыточно.»

**Why it was wrong.** I treated an authoring mistake in hand-written config as an
adversarial input. The list is small by design — it is typed into an environment
variable by one operator — so a duplicate is a typo that is visible on sight, and
enforcing it adds an error path and a message for a case the deployment model
itself excludes. The migration to some other user source is the moment that
validation becomes worth writing, not now.

**Rule.** "It could happen" is not enough to call a validation gap a bug when the
input is hand-written config of a bounded size. Report it as a judgment call, and
never block the docs on it.

**Check.** `grep -n "result\[token\]" internal/repositories/users/users.go`
shows a plain assignment with no duplicate branch — that is the accepted design,
not a missing check.
