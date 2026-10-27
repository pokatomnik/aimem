package qdrant

import "encoding/json"

// The Qdrant schema: the point ID and the request and response payloads of the
// four operations. Shapes follow the OpenAPI document of the server version
// pinned in docker-compose.yaml.

type (
	// envelope is the response Qdrant wraps around every result. The shape of
	// status differs by outcome: a successful answer carries the string "ok", a
	// failure carries an object with the server's description in error. WithStatus
	// reads both without failing on either.
	envelope[T any] struct {
		Result *T              `json:"result"`
		Status json.RawMessage `json:"status"`
	}

	collectionExistence struct {
		Exists bool `json:"exists"`
	}

	createCollectionRequest struct {
		Vectors map[string]vectorParams `json:"vectors"`
	}

	vectorParams struct {
		Size     uint32 `json:"size"`
		Distance string `json:"distance"`
		Datatype string `json:"datatype,omitempty"`
	}

	pointStruct struct {
		ID string `json:"id"`
		// The collection declares a named vector, so a point must name it too: a
		// bare array would address Qdrant's default vector "", which this
		// collection does not have.
		Vector  map[string][]float32 `json:"vector"`
		Payload map[string]any       `json:"payload"`
	}

	upsertPointsRequest struct {
		Points []pointStruct `json:"points"`
	}

	queryPointsRequest struct {
		Query       []float32 `json:"query"`
		Using       string    `json:"using"`
		Limit       int       `json:"limit"`
		WithPayload []string  `json:"with_payload"`
		WithVector  bool      `json:"with_vector"`
	}

	queryPointsResponse struct {
		Points []scoredPoint `json:"points"`
	}

	scoredPoint struct {
		ID      pointID        `json:"id"`
		Score   float32        `json:"score"`
		Payload map[string]any `json:"payload"`
	}

	retrievePointsRequest struct {
		IDs         []string `json:"ids"`
		WithPayload any      `json:"with_payload"`
		WithVector  bool     `json:"with_vector"`
	}

	retrievedPoint struct {
		ID      pointID        `json:"id"`
		Payload map[string]any `json:"payload"`
	}

	// deletePointsRequest is the point selector the delete endpoint accepts as
	// its body.
	deletePointsRequest struct {
		Points []string `json:"points"`
	}

	// pointID accepts both shapes Qdrant allows for a point ID: an unsigned
	// integer or a UUID string. IDs written by this package are UUIDs, but the
	// response is decoded uniformly so an unexpected shape is a clear error
	// rather than a silently empty field.
	pointID struct {
		UUID  string
		IsNum bool
	}
)

func (p *pointID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	if data[0] == '"' {
		var uuid string
		if err := json.Unmarshal(data, &uuid); err != nil {
			return err
		}

		p.UUID = uuid

		return nil
	}

	var num uint64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}

	p.IsNum = true
	p.UUID = formatNumericID(num)

	return nil
}
