package qdrant

// Each user's collection keeps the schema of the xvec store: the same payload keys
// and a single float32 vector named "embedding" compared by cosine similarity. The
// schema is identical for every user; only the collection name differs.
const (
	keyFieldSummary   = "summary"
	keyFieldFact      = "fact"
	keyFieldEmbedding = "embedding"
	keyFieldCreatedAt = "created_at"

	distanceCosine  = "Cosine"
	datatypeFloat32 = "float32"

	headerAPIKey      = "api-key"
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"

	queryParamWait = "wait"
)
