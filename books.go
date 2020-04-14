package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/go-chi/render"
)

// BookPayload is the request-response payload for a Book.
type BookPayload struct {
	*Book
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
func NewBookResponse(book *Book) *BookPayload {
	response := &BookPayload{Book: book}
	return response
}

// NewBookListResponse works similar to NewBookResponse, but for a slice.
func NewBookListResponse(books []*Book) []render.Renderer {
	list := []render.Renderer{}
	for _, book := range books {
		list = append(list, NewBookResponse(book))
	}
	return list
}

// Book data model.
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// Fixture data.
//
// TODO(simone): convert to actual database.
var books = []*Book{
	{ID: "1", Title: "Cattedrale", Author: "Carver"},
	{ID: "2", Title: "Uno, nessuno, centomila", Author: "Luigi Pirandello"},
}

func dbNewBook(book *Book) (string, error) {
	book.ID = fmt.Sprintf("%d", rand.Intn(100)+10)
	books = append(books, book)
	return book.ID, nil
}

func dbGetBook(id string) (*Book, error) {
	for _, b := range books {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, errors.New("book not found")
}

func dbUpdateBook(id string, book *Book) (*Book, error) {
	for i, b := range books {
		if b.ID == id {
			books[i] = book
			return book, nil
		}
	}
	return nil, errors.New("book not found")
}

func dbRemoveBook(id string) (*Book, error) {
	for i, b := range books {
		if b.ID == id {
			books = append(books[:i], books[i+1:]...)
			return b, nil
		}
	}
	return nil, errors.New("book not found")
}
