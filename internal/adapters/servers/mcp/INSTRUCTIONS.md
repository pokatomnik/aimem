# AI Memory MCP Server

This server gives the assistant access to aimem's persistent memory for facts and knowledge about the user. It can save a fact, find related memories by meaning, read a saved fact in full, and delete one. Use it for durable, useful context—not as a transcript or a general code-search tool.

## When to use memory

These are common triggers, not an exhaustive list or a restriction. Use the tools in other situations too whenever you judge that doing so would help preserve useful context or answer the user:

- **The user explicitly asks to remember or recall something.** Save the stated information, or search memory to answer a recall request.
- **The user shares critically important information.** Decide whether it is important enough to be useful in future conversations and save it.
- **The user shares personal information about themselves that is unrelated to code.** Save useful personal facts or preferences even if they did not explicitly ask.

Save only information the user actually provided; do not invent or silently embellish facts. Use judgment and avoid saving incidental details that are unlikely to be useful later.

Examples:

- “Remember that I prefer concise answers.” → call `remember`.
- “What dates did I say I would be in Berlin?” → call `recall` with a natural-language query.
- “I’m allergic to peanuts.” → call `remember` as personal information.

## Tools

- **`remember`** — Save a fact or knowledge. Pass the fact as `fact`; the result is a short summary. Use for explicit save requests and information that meets the triggers above.
- **`recall`** — Semantic search over saved memories using a natural-language `query`. Returns only the **top-K most relevant** matches, ranked by similarity, each with an ID, creation time, and a short **summary** — never the full text. Use it when the user asks what is remembered or when relevant past context is needed.
- **`read_fact`** — Read one saved fact in full by its `id`, usually from a `recall` result. Use this whenever a summary is too short or too vague to answer from; the response contains the complete stored text as well as the summary.
- **`remove_fact_by_id`** — Permanently delete one saved fact by its `id`, usually from a `recall` result. Use only when the user explicitly asks to forget something, or when a stored fact is wrong and the user agreed to remove it.

## Recall returns relevance, not inventory

The memory has no tool that lists everything it holds, and this is deliberate: the dataset is unbounded and dumping it would flood the context. `recall` is a relevance search, so it returns a capped set of best matches and nothing more. Two consequences to keep in mind:

- **A short or empty result does not mean memory is empty.** It means nothing scored high enough for that particular query. If the user asks what is stored and the results look thin, say that these are the most relevant matches rather than claiming it is the full picture.
- **Do not infer completeness from recall.** Never tell the user that a list of `recall` results is everything that is saved. Rephrase queries and recall again when you suspect more exists.

The `summary` field exists to keep `recall` cheap: full facts can be long, and returning many of them at once would crowd out the actual conversation. The summary is a compressed stand-in, not the record itself, and summarization can drop details. When exact wording, names, numbers, paths, or configuration values matter, treat the summary as a pointer and call `read_fact`.

If recall finds no relevant memory, say so rather than implying that the information was saved. Never claim a fact was remembered or retrieved unless the corresponding tool call succeeded.

## When to read more than one fact

A `recall` result is a ranked list, and several entries can be relevant at once. Call `read_fact` for each entry whose summary you actually intend to rely on, not just the first one:

- **The summaries overlap or seem to contradict each other.** One may be older or more specific; read both before you state anything as fact.
- **The question asks about history or change over time.** Summaries carry dates but drop nuance, so read the entries spanning the period in question.
- **The answer needs to be assembled from several partial summaries.** Read all of the ones you plan to combine.
- **A summary mentions something the user's question depends on but does not spell out.** Read it before filling the gap yourself.

Don't read every match reflexively — that defeats the point of summaries. Read the ones you will quote, cite, compare, or build on.
