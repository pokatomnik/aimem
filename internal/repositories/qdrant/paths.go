package qdrant

import (
	"fmt"
	"strings"
)

// collectionName maps a user to the Qdrant collection that holds their facts.
//
// Storage is partitioned per user: each user owns a whole collection rather than a
// slice of a shared one. The point ID is a pure function of the fact's text, so
// two users storing the same text would address the same point in a shared
// collection and the second AddFact would be answered with
// apperrors.ErrAlreadyHasFact. Separate collections make that collision
// impossible, and a fact belonging to someone else simply does not exist in the
// caller's collection.
//
// An empty userID resolves to the bare base name, so a single-user deployment and
// data written before users existed keep working against one collection.
//
// The name is only trimmed, not otherwise sanitized: Qdrant accepts ASCII letters,
// digits, "_" and "-", so a userID containing anything else is rejected by the
// server, and a userID containing "/" would address a different route entirely.
// Callers are expected to pass an opaque identifier, not a free-form string such
// as an email address.
func collectionName(userID string) string {
	const base = "memory"

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return base
	}

	return strings.Join([]string{base, userID}, "-")
}

// pathCollection is the collection root for a user, the base of every other path.
func pathCollection(userID string) string {
	return fmt.Sprintf("/collections/%s", collectionName(userID))
}

func pathCollectionExists(userID string) string {
	return fmt.Sprintf("%s/exists", pathCollection(userID))
}

func pathPoints(userID string) string {
	return fmt.Sprintf("%s/points", pathCollection(userID))
}

func pathPointsQuery(userID string) string {
	return fmt.Sprintf("%s/query", pathPoints(userID))
}

func pathPointsDelete(userID string) string {
	return fmt.Sprintf("%s/delete", pathPoints(userID))
}
