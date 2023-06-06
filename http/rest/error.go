package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/thunder33345/bookstore"
)

// ErrResponse is the response that get sent when an error occurs
type ErrResponse struct {
	//Err is for internal logic handling
	Err error `json:"-"`
	//HTTPStatusCode is for internal logic handling
	HTTPStatusCode int `json:"-"`

	//MessageText is a short error that get sent to the client
	MessageText string `json:"message"`
	//ErrorText is the full error chain for debugging
	ErrorText string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest creates a new generic invalid request response
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		MessageText:    "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ErrQueryResponse creates an error response, which attempts to handle common error types from querying
func ErrQueryResponse(err error) render.Renderer {
	e := &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		MessageText:    "Unhandled error.",
		ErrorText:      err.Error(),
	}

	//we check for common error types and set a predefined status code for these
	var noResErr *bookstore.NoResultError
	if errors.As(e.Err, &noResErr) {
		e.HTTPStatusCode = http.StatusNotFound
		e.MessageText = noResErr.Error()
	}

	var dupeErr *bookstore.DuplicateError
	if errors.As(e.Err, &dupeErr) {
		e.HTTPStatusCode = http.StatusBadRequest
		e.MessageText = dupeErr.Error()
	}

	var depErr *bookstore.DependedError
	if errors.As(e.Err, &depErr) {
		e.HTTPStatusCode = http.StatusConflict
		e.MessageText = depErr.Error()
	}

	var invDepErr *bookstore.InvalidDependencyError
	if errors.As(e.Err, &invDepErr) {
		e.HTTPStatusCode = http.StatusBadRequest
		e.MessageText = invDepErr.Error()
	}

	var invID *bookstore.NonExistentIDError
	if errors.As(e.Err, &invID) {
		e.HTTPStatusCode = http.StatusNotFound
		e.MessageText = invID.Error()
	}

	return e
}

func ErrInvalidRequestBody(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		MessageText:    "Invalid request body, error while parsing request body.",
		ErrorText:      err.Error(),
	}
}

func ErrInvalidRequestParam(param string, err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		MessageText:    fmt.Sprintf("Invalid request parameter(%s).", param),
		ErrorText:      err.Error(),
	}
}

// ErrInvalidIDRequest creates a new invalid request response due to invalid UUID
func ErrInvalidIDRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		MessageText:    "Invalid ID.",
		ErrorText:      err.Error(),
	}
}

func ErrProcessingFile(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		MessageText:    "Error processing file.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		MessageText:    "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, MessageText: "Resource not found."}

var ErrUnauthorized = &ErrResponse{HTTPStatusCode: http.StatusUnauthorized, MessageText: "Unauthorized."}
var ErrForbidden = &ErrResponse{HTTPStatusCode: http.StatusForbidden, MessageText: "Forbidden."}
