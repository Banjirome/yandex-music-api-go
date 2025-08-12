package models

// Response универсальная обёртка ответа API.
type Response[T any] struct {
	InvocationInfo *InvocationInfo `json:"invocationInfo,omitempty"`
	Result         T               `json:"result"`
	Pager          *Pager          `json:"pager,omitempty"`
}

type InvocationInfo struct {
	ReqID string `json:"req-id,omitempty"`
}

type Pager struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
	Total   int `json:"total"`
}

// Revision represents structure with single revision field (e.g., library track like/dislike operations).
type Revision struct {
	Revision int `json:"revision"`
}

type APIError struct {
	InvocationInfo *InvocationInfo `json:"invocationInfo,omitempty"`
	Body           *ErrorBody      `json:"error,omitempty"`
	StatusCode     int             `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Body != nil && e.Body.Message != "" {
		return e.Body.Message
	}
	return "yandex music api error"
}

// ErrorBody returns the underlying API error body (compat alias to C# field name "Error").
func (e *APIError) ErrorBody() *ErrorBody {
	if e == nil {
		return nil
	}
	return e.Body
}

type ErrorBody struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
