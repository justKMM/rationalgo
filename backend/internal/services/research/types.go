package research

// ApiResponse is the standard envelope every research endpoint returns.
type ApiResponse[T any] struct {
	Data       *T           `json:"data"`
	Metadata   EndpointMeta `json:"metadata"`
	Confidence float64      `json:"confidence"`
	Error      *ApiError    `json:"error,omitempty"`
}

// ApiError describes a request-level failure (e.g. missing/invalid input).
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Provenance fields embedded in every payload's data.
type Provenance struct {
	Source      string `json:"source"`
	RetrievedAt string `json:"retrieved_at"`
}
