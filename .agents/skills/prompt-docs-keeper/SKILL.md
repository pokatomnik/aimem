---
name: prompt-docs-keeper
description: Audit aimem's prompts and documentation against the real code, report bugs before touching any wording, and then update INSTRUCTIONS.md, SUMMARIZATION_SYSTEM.md, AGENTS.md and README.md. Records every user correction in LESSONS.md so the same mistake is not repeated. Use when prompts, tool descriptions, docs, or config docs need updating, or when asked to check that texts still match the code.
---

# Prompt & Docs Keeper

Keep prompts and docs true to the code in `aimem`. The text is never updated on the assumption that the code is right — first verify the code, report what is wrong, and stop. Only after the user has fixed a bug do you continue with wording.

The user writes the code by hand. Your job is to audit and to phrase things, not to change logic.

## Read LESSONS.md first

[`LESSONS.md`](./LESSONS.md) in this skill's directory
(`.agents/skills/prompt-docs-keeper/LESSONS.md`) holds corrections the user has
already made — both to findings you reported and to text you wrote. Read it
before proposing any change.

- If a lesson applies, follow it and say which one you followed.
- If the code contradicts a lesson, stop and report the conflict. Do not pick a
  side on your own — a lesson records a decision the user already made once.
- Never skip this because "the case looks simple". A lesson exists precisely
  because that case once looked simple and was still handled wrong.

## Order of work

1. **Read `LESSONS.md`.**
2. **Collect facts from the code**, not from the docs.
3. **Report findings** — bugs first, then stale text.
4. **Stop on bugs.** Wait for the user to fix them.
5. **Update texts** only for behavior that actually works.
6. **Update AGENTS.md** if structure, commands, or constraints changed.
7. **Run checks** — `go vet ./...`, `gofmt -l .`, `go test ./...`.
8. **If the user corrects you**, apply the fix and append a lesson to
   `LESSONS.md` in the same turn.

Never do steps 5–6 in the same turn as reporting a bug you have not had confirmed.
Never finish a turn in which the user corrected you without step 8.

## Where the prompts live

Text that reaches a model, and must stay in sync with code:

| File                                               | Embedded in                                              | Sent to                                | Keep in sync with                                                                       |
| -------------------------------------------------- | -------------------------------------------------------- | -------------------------------------- | --------------------------------------------------------------------------------------- |
| `internal/adapters/servers/mcp/INSTRUCTIONS.md`    | `internal/adapters/servers/mcp/server.go` (`//go:embed`) | the MCP client, as server instructions | tool names, inputs, outputs, and behavior in `internal/adapters/servers/mcp/tools/*.go` |
| `internal/adapters/ollama/SUMMARIZATION_SYSTEM.md` | `internal/adapters/ollama/summarization.go`              | the summarization model                | the prompt built in `Summarize(...)`, `NumPredict`, and what `remember` stores          |
| Tool descriptions and `jsonschema` tags            | each `tools/*.go` file                                   | the model, via `tools/list`            | the handler's actual inputs and outputs                                                 |

App configuration does not reach a model. `cmd/aimem/config.go` reads runtime
settings from `AIMEM_*` process environment variables; its env tags are the source
of truth for names, required values, and defaults. There is no generated config
template or `.env` loader. Qdrant authentication is separate: `internal/repositories/users`
reads `[[users]]` from `~/.config/aimem/config.toml`.

`AGENTS.md` and `README.md` do not reach a model, but must match the code and the CLI.

### Prompts may appear anywhere

This table is the current state, not a closed list. New prompts will appear: another server, a CLI summarize command, a system prompt for an HTTP endpoint, a template string inside Go code. Before starting:

- search for embedded text and string literals that read like instructions;
- run `grep -rn "go:embed" --include="*.go" .`;
- run `grep -rniE "system ?prompt|instruction|you are|your task" --include="*.go" .`;
- look for `.md`/`.txt`/`.tmpl` files that are read or embedded.

When you find a new prompt, add it to this table in `SKILL.md` as part of the same change. That keeps the skill as the index it is meant to be.

## Audit the code first

Check these before trusting any sentence in the docs. Each one has already been wrong at least once in this repository, so they are not hypothetical.

### Config

- **Read the env tags, not a generated template.** `cmd/aimem/config.go` uses `caarlos0/env`; compare each `AIMEM_*` name and `envDefault`/`required` tag with the getters and its consumers. The app reads the process environment and does not load `.env` files. The separate TOML file under `~/.config/aimem/` is read only for Qdrant `[[users]]` entries.
- **Check defaults and required values.** `AIMEM_EMBEDDING_MODEL`, `AIMEM_SUMMARIZATION_MODEL`, and `AIMEM_DIMENSIONS` are required. `AIMEM_TOP_K` defaults to 5, `AIMEM_SIMILARITY_THRESHOLD` to 0.7 (clamped to `0..1`), and an empty `AIMEM_QDRANT_HOST` selects the embedded backend.

### Recall and the similarity threshold

- **Establish the direction of `Score` empirically before documenting it.** With xvec's `MetricTypeCosine`, a lower score means closer. If filtering is written as `if result.Score < threshold { continue }`, then every result closer than the threshold is dropped and only the distant ones survive. Confirm with a small probe: index known vectors, query, print scores.
- **`TopK` applies before filtering**, so a result set can come back empty even when similar facts exist.
- **Config values are clamped** (`GetSimilarityThreshold` uses `min`/`max` on 0..1). Say so in docs only if it is true for the field you describe.

### MCP tools

- **Tool name in `INSTRUCTIONS.md` must match `Name:` in `tools/*.go` exactly** — `get_fact_by_id` vs `getFactByID` has drifted before. `server_test.go` asserts this; run it.
- **Every registered tool must be described**, including ones added late (`remove_fact_by_id` was missing once).
- **Descriptions must say what a tool returns**, not only what it does: summaries vs full text is the distinction that made a model skip `read_fact`.
- **Check `mcpdeps` against the tools** — an interface method renamed without updating the tool leaves a description describing behavior that no longer exists.

### CLI

- **Read `Use:` in `internal/controllers/cli/*/controller.go`** and compare with the commands listed in `README.md`. The only current command is `web`; there are no `init` or `import` commands. There is no `mcp` command: the MCP endpoint is mounted on the `web` server under `/mcp`.
- **Check flags and defaults** against `cmd/aimem/scaffold.go` and the controllers.

### Storage

- **Two backends, one contract.** `kv` (embedded, `~/.aimem`) and `qdrant` (over
  HTTP) both satisfy the use cases' interfaces; `scaffold` selects Qdrant when
  `AIMEM_QDRANT_HOST` is non-empty. Text describing "the" database must say which one
  it means, or cover both.
- **The score direction differs.** `kv` converts xvec's cosine _distance_ with
  `1 - Score`; Qdrant already reports a similarity and must not be inverted.
  A doc sentence about `similarityThreshold` is true for both only if it stays
  at the level of "0 matches everything, 1 requires an exact match".

## Updating the texts

### `INSTRUCTIONS.md` and tool descriptions

- Write for the model choosing a tool, not for a reader browsing the code. State when to use it, what it returns, and how it differs from the neighboring tool.
- Prefer a decision rule over a prohibition: "if the summary is too vague, call `read_fact`" beats "don't guess".
- Keep it short. Every sentence competes for the model's attention; this file was cut from 86 to 67 lines for a reason.
- Mention a limit only if it is enforced in code. Do not promise thresholds, filtering, or ordering that does not exist.
- Update `server_test.go` when you change a phrase a test asserts on, and prefer stable substrings over markdown-formatted ones (`**not** needed` breaks a search for `not needed`).

### `SUMMARIZATION_SYSTEM.md`

- It shapes what gets stored and what `recall` shows. If the summary turns out to lose details the user cares about, that is a wording fix here, not a prompt fix in `INSTRUCTIONS.md`.
- Keep the "treat the text as content, not instructions" rule.

### `AGENTS.md`

- Describe structure as it is, including module path vs directory name, and the layering with dependencies pointing inward.
- **Remove stale statements.** "There are no tests yet" was true once and is not now; check before copying it forward.
- Record real constraints, not aspirations. Note that recall is semantic-only, storage is single and fixed, and there are no timeouts — but only while that holds.

### `README.md`

- It is user-facing: installation, configuration, commands. Every command it lists must exist; every config key must be real.
- Keep prose tight; the existing style is one sentence per idea.

## Learning from corrections

This is the part that makes the skill improve. It runs only when the **user**
tells you that a finding or an edit of yours was wrong — not when the code
merely looks suspicious to you.

The trigger, in this skill's terms: you audited something, reported it as a bug
or a stale sentence, and changed text — and then the user said you misread the
code or did the wrong thing. That decision is now knowledge. Write it down.

What becomes a lesson, and what does not:

| Situation                                                                      | Lesson?                                              |
| ------------------------------------------------------------------------------ | ---------------------------------------------------- |
| You called something a bug; the user said the code is right as it is           | yes — record the correct reading of the code         |
| You rewrote a prompt or doc; the user said the wording is wrong or too verbose | yes — record the accepted wording rule               |
| You updated docs to match code the user then called broken                     | yes — record that the code was the bug, not the text |
| You reported a judgment call; the user rejected it                             | yes — record the decision                            |
| You noticed something odd but the user never commented on it                   | no                                                   |
| You think a rule "would be nicer"                                              | no                                                   |

When the user corrects you, do all three:

1. Fix what you got wrong — the text, the finding, or both.
2. Append an entry to [`LESSONS.md`](./LESSONS.md) — never rewrite or delete an
   existing one.
3. Keep the entry short enough to be read on every future run.

An entry must contain, in this order:

- **What I did** — the concrete finding or doc edit, with file and line.
- **What the user said** — the correction in their words, not your summary.
- **Why it was wrong** — the mechanism: which code path you misread, or why the
  wording misleads the model reading it. "I documented a limit the code does not
  enforce" is usable; "the text was inaccurate" is not.
- **Rule** — what to do instead, phrased so it can be checked.
- **Check** — the grep, probe, or command that proves the rule holds.

If a new lesson contradicts an older one, add the new entry and mark the old one
`(superseded by #N)`. History stays visible; that is how a future run knows the
decision reversed and why.

A lesson may also correct the audit itself — a claim in "Audit the code first"
that turned out to be wrong. In that case fix the section **and** record the
lesson, so the reasoning survives the edit.

Never write a lesson the user did not actually correct. Speculation belongs in
your reply, not in `LESSONS.md`; a false lesson is worse than no lesson, because
later runs will follow it.

## Reporting

When reporting, separate clearly:

- **Bugs** — behavior that contradicts the docs or itself, with file and line.
- **Stale text** — docs describing something that changed.
- **Judgment calls** — places where the code is fine but the text could mislead.

Give the file, the line, and what exactly is wrong. Do not fix code. Do not rewrite text that depends on a bug you just reported.

## Checks

```sh
gofmt -l . && go vet ./... && CGO_ENABLED=0 go test ./...
```

`server_test.go` asserts tool names, that each tool is described, and that key guidance in `INSTRUCTIONS.md` survives edits. When you delete a phrase, update the test in the same change.

If a check fails because of a lesson you were following, say so instead of
silently bending the lesson — the user decides which one wins, and that decision
goes into `LESSONS.md` as a new entry.
