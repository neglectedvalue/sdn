package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-redis/cache/v8"
)

type Server struct {
	RedisCache *cache.Cache
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("You posted to /."))
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
