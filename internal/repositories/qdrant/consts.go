package qdrant

// The collection keeps the schema of the xvec store: the same payload keys and a
// single float32 vector named "embedding" compared by cosine similarity.
const (
	collectionName = "memory"

	keyFieldSummary   = "summary"
	keyFieldFact      = "fact"
	keyFieldEmbedding = "embedding"
	keyFieldCreatedAt = "created_at"

	pathCollection       = "/collections/" + collectionName
	pathCollectionExists = pathCollection + "/exists"
	pathPoints           = pathCollection + "/points"
	pathPointsQuery      = pathPoints + "/query"
	pathPointsDelete     = pathPoints + "/delete"

	distanceCosine  = "Cosine"
	datatypeFloat32 = "float32"

	headerAPIKey      = "api-key"
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"

	queryParamWait = "wait"
)
