package presenter

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/BuriedInTheGround/book-api-example/data"
	"github.com/go-chi/render"
)

type apiContextKey string

// BookPayload is the request-response payload for a Book.
type BookPayload struct {
	*data.Book
}

// Bind makes BookPayload a render.Binder.
func (bp *BookPayload) Bind(r *http.Request) error {
	fmt.Println(r.Context().Value(apiContextKey("book")))
	if bp.Book == nil {
		return errors.New("missing required Book fields")
	}
	return nil
}

// Render makes BookPayload a render.Renderer.
func (bp *BookPayload) Render(w http.ResponseWriter, r *http.Request) error {
	// Do pre-processing here..
	return nil
}

// NewBookResponse generates a BookPayload response for the given book.
func NewBookResponse(book *data.Book) *BookPayload {
	response := &BookPayload{Book: book}
	return response
}

// NewBookListResponse works similar to NewBookResponse, but for a slice.
func NewBookListResponse(books []*data.Book) []render.Renderer {
	list := []render.Renderer{}
	for _, book := range books {
		list = append(list, NewBookResponse(book))
	}
	return list
}
