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
