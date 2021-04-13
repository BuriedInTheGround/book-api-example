package presenter

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrorResponse is a payload for errors handling.
type ErrorResponse struct {
	Error          error `json:"-"` // Low-level runtime error.
	HTTPStatusCode int   `json:"-"` // HTTP response status code.

	StatusText string `json:"status"`          // User-level status message.
	AppCode    int64  `json:"code,omitempty"`  // Application-specific error code.
	ErrorText  string `json:"error,omitempty"` // Application-level error message used for debugging.
}

// Render makes ErrorResponse a render.Renderer.
func (rd *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, rd.HTTPStatusCode)
	// Set the application-specific error code here..
	return nil
}

// ErrorInvalidRequest return a Renderer for invalid request type errors.
func ErrorInvalidRequest(err error /*, appCode int64*/) render.Renderer {
	return &ErrorResponse{
		Error:          err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid Request.",
		ErrorText:      err.Error(),
		/*AppCode: appCode,*/
	}
}

// ErrorRender return a Renderer for render-type errors.
func ErrorRender(err error) render.Renderer {
	return &ErrorResponse{
		Error:          err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// ErrorNotFound is a Renderer for resource not found errors.
var ErrorNotFound = &ErrorResponse{
	HTTPStatusCode: 404,
	StatusText:     "Resource not found.",
}

// ErrorInternalServer is a Renderer for internal server errors.
var ErrorInternalServer = &ErrorResponse{
	HTTPStatusCode: 500,
	StatusText:     "Internal server error.",
}
