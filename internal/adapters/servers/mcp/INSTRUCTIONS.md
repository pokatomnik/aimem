# AI Memory MCP Server

Persistent memory of facts and knowledge about the user. Save a fact, find related memories by meaning, read one in full, or delete one. Not a transcript, not a code search.

## When to use memory

Common triggers, not a restriction — use judgment:

- **The user asks to remember or recall something.** Save it or search for it.
- **The user shares important information** worth having in future conversations.
- **The user shares personal facts or preferences.** Save them even without an explicit request.
- **The message assumes context you lack.** Recall first — see below.

Save only what the user actually provided; never invent or embellish.

## When to recall without being asked

Ask yourself: **does this message assume knowledge I lack — neither in this conversation nor in the repository?**

Recall when:

- **It refers to shared history.** “as we agreed”, “again”, “last time”, “the usual”.
- **It names a project or system you cannot locate** in the current context.
- **It states a preference or working style.** “keep it short”, “same style as before”.
- **It is the first substantive message of a conversation**, about the user, their setup, or past work.
- **It claims something about your past behaviour.** “you said”, “you recommended”.

Don't recall when:

- **The message answers itself** — a self-contained question, an error with a stack trace.
- **It is a general question about programming**, not about this user. Memory holds facts, not docs.
- **The context is already established**, unless the topic shifts.
- **Work is in progress** — files and conversation are the source of truth.

These limits on contextual recall do not apply to the conflict check after saving new information.

A few calls per conversation is normal. Recall on **every message** is not: each search is a round trip, and irrelevant memories crowd out the answer.

## Memory is the weakest source of truth

Current message > repository > earlier in this conversation > memory.

Stored facts can be stale or wrong. If memory contradicts the code, say so rather than quietly following one. A weak match is not evidence that the user's premise is wrong — treat it as a lead, not a conclusion.

## Tools

- **`remember`** — save a fact; returns a short summary. That summary is what `recall` will later show, so write a self-contained statement.
- **`recall`** — semantic search by natural-language `query`. Returns only the **top-K most relevant** matches with ID, date, and a short summary — never the full text.
- **`read_fact`** — read one fact in full by `id` from `recall`. Use when a summary is too vague: exact wording, names, numbers, paths.
- **`remove_fact_by_id`** — permanently delete a fact only when the user explicitly requests its deletion. Confirming that it is wrong is not permission to delete it.

## Conflicting or outdated facts

When saving new information, follow this order even if the current context is already established:

1. Save the new fact with `remember` first. Do not delete an older fact to make room for it.
2. After a successful save, use `recall` to find related older facts. Use `read_fact` to check their full contents for a full or partial conflict; similarity or a short summary alone is not proof.
3. For every confirmed conflict, proactively show the user the old fact's ID and relevant contents, explain the conflicting part and any still-valid information that deleting the whole fact would also remove, and ask whether to delete the old fact. Do this even if the user has forgotten the old fact exists.
4. Call `remove_fact_by_id` only after the user explicitly requests its deletion, including explicitly confirming deletion of the identified fact. Agreeing that a fact is wrong is not permission to delete it. Declining or not answering leaves both facts stored.

If the user already explicitly asked to delete the identified fact, carry out that request without asking again. Never infer permission from a correction, a newer fact, or a contradiction.

## Recall returns relevance, not inventory

There is no tool that lists everything, deliberately: the set is unbounded and dumping it floods the context. `recall` returns the closest matches above a relevance threshold — at most a configured cap. So a short or empty result **does not mean memory is empty**: either nothing scored above the threshold, or the handful shown were simply the closest of many. Never present recall results as the complete picture; rephrase and search again if you suspect more.

The `summary` field keeps `recall` cheap. It is a compressed pointer, not the record, and summarization drops details — follow it with `read_fact` when precision matters.

If nothing relevant is found, say so. Never claim a fact was remembered or retrieved unless the tool call succeeded.

## When to read more than one fact

Read every entry whose summary you intend to rely on, not just the first:

- **Summaries overlap or contradict** — read both before stating anything as fact.
- **The question is about history** — read the entries spanning that period.
- **The answer combines several partial summaries.**
- **A summary hints at something your answer depends on** but does not spell out.

Don't read every match reflexively — only the ones you will quote, compare, or build on.
