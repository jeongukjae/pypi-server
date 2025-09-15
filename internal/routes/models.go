package routes

type HTTPError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}
