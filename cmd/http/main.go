package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Headers  string `json:"headers"`
	Body     string `json:"body"`
	Filename string `json:"filename"`
}

type Payload struct {
	ID  int    `json:"id"`
	Msg string `json:"msg"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var headerValue string

		for key, values := range r.Header {
			headerValue += fmt.Sprintf("%s: %s,", key, strings.Join(values, ", "))
		}
		resp := Response{
			Headers: headerValue,
		}

		j, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		w.Write(j)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		r.Body.Read([]byte(""))

		var headerValue string

		for key, values := range r.Header {
			headerValue += fmt.Sprintf("%s: %s,", key, strings.Join(values, ", "))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonStr := string(body)

		filename := ""
		f, h, _ := r.FormFile("file")
		if h != nil {
			fmt.Println(h)
			filename = string(h.Size)
			defer f.Close()
		}

		resp := Response{
			Headers:  headerValue,
			Body:     jsonStr,
			Filename: filename,
		}

		j, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		w.Write(j)
	})

	r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var headerValue string

		for key, values := range r.Header {
			headerValue += fmt.Sprintf("%s: %s,", key, strings.Join(values, ", "))
		}
		resp := Response{
			Headers: headerValue,
		}

		j, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		w.Write(j)
	})

	r.Put("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var headerValue string

		for key, values := range r.Header {
			headerValue += fmt.Sprintf("%s: %s,", key, strings.Join(values, ", "))
		}
		resp := Response{
			Headers: headerValue,
		}

		j, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		w.Write(j)
	})

	http.ListenAndServe(":3000", r)
}
