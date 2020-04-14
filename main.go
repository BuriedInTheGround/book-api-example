package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/BuriedInTheGround/book-api-example/data"
	"github.com/BuriedInTheGround/book-api-example/presenter"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

type apiContextKey string

var (
	mysqlHost     = os.Getenv("MYSQL_HOST")
	mysqlUser     = os.Getenv("MYSQL_USER")
	mysqlPassword = os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase = os.Getenv("MYSQL_DB")
	dbSession     sqlbuilder.Database
)

func main() {
	dbSession = initDatabase()

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

func initDatabase() sqlbuilder.Database {
	sess, err := mysql.Open(mysql.ConnectionURL{
		User:     mysqlUser,
		Password: mysqlPassword,
		Host:     mysqlHost,
		Database: mysqlDatabase,
	})
	if err != nil {
		log.Fatalln("Cannot connect to the database.")
	}
	_, err = sess.Exec(
		`CREATE TABLE IF NOT EXISTS books (
			id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT, PRIMARY KEY(id),
			title VARCHAR(255),
			author VARCHAR(255)
		)`,
	)
	if err != nil {
		log.Fatalln("[Error] Cannot initialize the database.")
	}
	if os.Getenv("CLEAR_DB_ON_RELOAD") != "" && os.Getenv("CLEAR_DB_ON_RELOAD") == "on" {
		if err = sess.Collection("books").Truncate(); err != nil {
			log.Println("[Warning] Cannot truncate table 'books'.")
		}
	}
	return sess
}

// ListBooks returns all Book items stored.
func ListBooks(w http.ResponseWriter, r *http.Request) {
	var books []*data.Book
	dbSession.Collection("books").Find().All(&books)
	if err := render.RenderList(w, r, presenter.NewBookListResponse(books)); err != nil {
		render.Render(w, r, presenter.ErrorRender(err))
		return
	}
}

// BookCtx is a middleware that retrieve a Book from the ID in the URL.
func BookCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var book *data.Book
		var bookID int64
		var err error

		bookIDStr := chi.URLParam(r, "bookID")
		bookID, err = strconv.ParseInt(bookIDStr, 10, 32)
		if err != nil {
			render.Render(w, r, presenter.ErrorInvalidRequest(err))
			return
		}
		book, err = dbGetBook(int(bookID))

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
	bookID, err := dbNewBook(book)
	if err != nil {
		render.Render(w, r, presenter.ErrorInternalServer)
	}

	book.ID = bookID

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
	newBook, err := dbUpdateBook(book.ID, book)
	if err != nil {
		render.Render(w, r, presenter.ErrorInternalServer)
		return
	}

	render.Render(w, r, presenter.NewBookResponse(newBook))
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

func dbNewBook(book *data.Book) (int, error) {
	id, err := dbSession.Collection("books").Insert(book)
	return int(id.(int64)), err
}

func dbGetBook(id int) (*data.Book, error) {
	var book data.Book
	if err := dbSession.Collection("books").Find(id).One(&book); err == nil {
		return &book, nil
	}
	return nil, errors.New("book not found")
}

func dbUpdateBook(id int, book *data.Book) (*data.Book, error) {
	err := dbSession.Collection("books").Find(int64(id)).Update(*book)
	if err != nil {
		return nil, errors.New("book not found")
	}
	return book, nil
}

func dbRemoveBook(id int) (*data.Book, error) {
	var book data.Book
	res := dbSession.Collection("books").Find(id)
	if err := res.One(&book); err == nil {
		res.Delete()
		return &book, nil
	}
	return nil, errors.New("book not found")
}
