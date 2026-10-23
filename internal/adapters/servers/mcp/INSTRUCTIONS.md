# AI Memory MCP Server

This server gives the assistant access to aimem's persistent memory for facts and knowledge about the user. It can save a fact, find related memories by meaning, and retrieve a saved fact in full. Use it for durable, useful context—not as a transcript or a general code-search tool.

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
- **`recall`** — Find memories related to a natural-language `query`. Results contain an ID, creation time, and summary. Recall is semantic and returns summaries, not full records. Use it when the user asks what is remembered or when relevant past context is needed.
- **`getFactByID`** — Retrieve the full saved fact using its `id`, usually from a `recall` result. Use this when a summary does not provide enough detail; answer from the returned content.

If recall finds no relevant memory, say so rather than implying that the information was saved. Never claim a fact was remembered or retrieved unless the corresponding tool call succeeded.
