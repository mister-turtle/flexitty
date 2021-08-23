package webserver

import (
	"html/template"
	"log"
	"net/http"
)

type renderer struct {
	templates *template.Template
}

func (r renderer) RenderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 - Not found"))
}

func (r renderer) RenderBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 - Bad Request"))
}

func (r renderer) RenderInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error"))
}

func newRenderer() renderer {
	templates, err := template.ParseFS(templateFS, "embedded/*.html")
	if err != nil {
		log.Fatal(err)
	}

	return renderer{
		templates: templates,
	}
}
