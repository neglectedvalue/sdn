package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/google/uuid"
)

type Server struct {
	RedisCache *cache.Cache
	BaseURL    string
}

type Note struct {
	Data     []byte
	Destruct bool
}

func (s *Server) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method == "GET" || r.Method == "HEAD" {
		s.handleGet(w, r)
		return
	}

	if r.Method == "POST" && r.URL.Path == "/" {
		s.handlePost(w, r)
		return
	}

	s.notFound(w, r)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	mediaType := r.Header.Get("Content-Type")
	if mediaType != "application/x-www-form-urlencoded" {
		s.badRequest(w, r, http.StatusBadRequest, "Invalid media type posted.")
		return
	}

	err := r.ParseForm()
	if err != nil {
		s.badRequest(w, r, http.StatusBadRequest, "Invalid form data posted.")
		return
	}
	form := r.PostForm
	message := form.Get("message")
	destruct := false
	ttl := time.Hour * 24
	if form.Get("ttl") == "untilRead" {
		destruct = true
		ttl = ttl * 365
	}

	note := &Note{
		Data:     []byte(message),
		Destruct: destruct,
	}

	key := uuid.NewString()
	err = s.RedisCache.Set(
		&cache.Item{
			Ctx:            r.Context(),
			Key:            key,
			Value:          note,
			TTL:            ttl,
			SkipLocalCache: true,
		})
	if err != nil {
		fmt.Println(err)
		s.serverError(w, r)
		return
	}

	noteURL := fmt.Sprintf("%s/%s", s.BaseURL, key)
	w.WriteHeader(http.StatusOK)
	s.renderMessage(
		w, r,
		"Note was successfully created",
		template.HTML(
			fmt.Sprintf("<a href='%s'>%s</a>", noteURL, noteURL)))
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		s.renderTemplate(w, r, nil,
			"layout",
			"dist/layout.html",
			"dist/index.html")
		return
	}

	noteId := strings.TrimPrefix(path, "/")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		fmt.Sprintf(
			"You requested note with ID '%s'",
			noteId)))
}

func (s *Server) renderMessage(
	w http.ResponseWriter,
	r *http.Request,
	title string,
	paragraphs ...interface{},
) {
	s.renderTemplate(
		w, r,
		struct {
			Title      string
			Paragraphs []interface{}
		}{
			Title:      title,
			Paragraphs: paragraphs,
		},
		"layout",
		"dist/layout.html",
		"dist/message.html",
	)
}

func (s *Server) badRequest(w http.ResponseWriter, r *http.Request,
	statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Ops something went wrong. Please check the server logs."))
}

func (s *Server) renderTemplate(w http.ResponseWriter,
	r *http.Request,
	data interface{},
	name string,
	files ...string) {
	t := template.Must(template.ParseFiles(files...))
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		panic(err)
	}
}
