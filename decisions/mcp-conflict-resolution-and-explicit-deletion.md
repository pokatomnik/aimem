# MCP Conflict Resolution and Explicit Deletion Requests

## Decision

The MCP server instructions and tool descriptions direct the assistant to save a
new user-provided fact first. After a successful save, it uses semantic recall to
find related older facts and reads their full contents to check for full or
partial contradictions. A similar match or a compressed summary is not enough to
establish a contradiction.

For each confirmed conflict, the assistant proactively presents the old fact's ID
and relevant contents, explains the conflict and any still-valid information that
deleting the whole record would lose, and asks whether to delete it. The user
need not remember that the older fact exists.

Every deletion requires an explicit user request. Explicit confirmation of a
question about deleting an identified fact counts; merely correcting information
or agreeing that a fact is wrong does not. Until the user requests deletion, both
facts remain stored. A direct deletion request is carried out without asking for
the same permission again.

For this MCP workflow, saving the new fact before deleting an old one supersedes
the delete-then-write ordering described in
[Stored Records Are Immutable](stored-records-immutable.md). Records remain
immutable; neither backend's storage contract changes. Deletion remains
[idempotent](delete-is-idempotent.md).

These are instructions for the client assistant, not server-side conflict
detection or consent enforcement.

## Why

A newer fact can contradict only part of an older record. Automatically deleting
or replacing the old record would also discard information that is still valid,
and correcting information does not authorize that loss. Silently keeping both
without raising the conflict leaves the user unaware of contradictory memory.

Save-first, proactive disclosure and an explicit deletion request preserve the
new information while leaving the destructive decision with the user. Automatic
conflict cleanup and delete-before-save replacement are rejected.
