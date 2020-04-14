package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/BuriedInTheGround/book-api-example/data"
	"github.com/BuriedInTheGround/book-api-example/presenter"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type apiContextKey string

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Book API. Try some endpoints, like '/books'."))
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/books", func(r chi.Router) {
		r.Get("/", ListBooks)
		r.Post("/", CreateBook)

		r.Route("/{bookID}", func(r chi.Router) {
			r.Use(BookCtx)
			r.Get("/", GetBook)
			r.Put("/", UpdateBook)
			r.Delete("/", DeleteBook)
		})
	})

	http.ListenAndServe(":3000", r)
}

// ListBooks returns all Book items stored.
func ListBooks(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, presenter.NewBookListResponse(books)); err != nil {
		render.Render(w, r, presenter.ErrorRender(err))
		return
	}
}

// BookCtx is a middleware that retrieve a Book from the ID in the URL.
func BookCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var book *data.Book
		var err error

		bookID := chi.URLParam(r, "bookID")
		book, err = dbGetBook(bookID)

		if err != nil {
			render.Render(w, r, presenter.ErrorNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), apiContextKey("book"), book) //nolint
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateBook adds the posted Book item.
func CreateBook(w http.ResponseWriter, r *http.Request) {
	data := &presenter.BookPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, presenter.ErrorInvalidRequest(err))
		return
	}

	book := data.Book
	dbNewBook(book)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, presenter.NewBookResponse(book))
}

// GetBook returns a specific Book.
func GetBook(w http.ResponseWriter, r *http.Request) {
	book := r.Context().Value(apiContextKey("book")).(*data.Book)

	if err := render.Render(w, r, presenter.NewBookResponse(book)); err != nil {
		render.Render(w, r, presenter.ErrorRender(err))
		return
	}
}

// UpdateBook updates a Book item's data.
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	book := r.Context().Value(apiContextKey("book")).(*data.Book)

	data := &presenter.BookPayload{Book: book}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, presenter.ErrorInvalidRequest(err))
		return
	}

	book = data.Book
	dbUpdateBook(book.ID, book)

	render.Render(w, r, presenter.NewBookResponse(book))
}

// DeleteBook removes a specific Book.
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	var err error

	book := r.Context().Value(apiContextKey("book")).(*data.Book)

	book, err = dbRemoveBook(book.ID)
	if err != nil {
		render.Render(w, r, presenter.ErrorInvalidRequest(err))
		return
	}

	render.Render(w, r, presenter.NewBookResponse(book))
}

// Fixture data.
//
// TODO(simone): convert to actual database.
var books = []*data.Book{
	{ID: "1", Title: "Cattedrale", Author: "Carver"},
	{ID: "2", Title: "Uno, nessuno, centomila", Author: "Luigi Pirandello"},
}

func dbNewBook(book *data.Book) (string, error) {
	book.ID = fmt.Sprintf("%d", rand.Intn(100)+10)
	books = append(books, book)
	return book.ID, nil
}

func dbGetBook(id string) (*data.Book, error) {
	for _, b := range books {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, errors.New("book not found")
}

func dbUpdateBook(id string, book *data.Book) (*data.Book, error) {
	for i, b := range books {
		if b.ID == id {
			books[i] = book
			return book, nil
		}
	}
	return nil, errors.New("book not found")
}

func dbRemoveBook(id string) (*data.Book, error) {
	for i, b := range books {
		if b.ID == id {
			books = append(books[:i], books[i+1:]...)
			return b, nil
		}
	}
	return nil, errors.New("book not found")
}
