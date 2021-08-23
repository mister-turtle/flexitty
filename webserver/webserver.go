package webserver

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mister-turtle/flexitty/broker"
)

type Options struct {
	Address string
	Port    int
}

type WebServer struct {
	renderer     renderer
	socketBroker *broker.Broker
}

func Start(o Options) error {

	server := WebServer{
		renderer:     newRenderer(),
		socketBroker: broker.New(),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", server.IndexHandler)
	r.Get("/new", server.NewSessionHandler)
	r.Get("/session/{uuid}", server.TTYHandler)
	r.Get("/session/{uuid}/ws", server.WebSocketHandler)

	staticStub, err := fs.Sub(staticFS, "embedded")
	if err != nil {
		log.Fatal(err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticStub))))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", o.Address, o.Port), r)
}

func (ws WebServer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Working!"))
}

func (ws WebServer) NewSessionHandler(w http.ResponseWriter, r *http.Request) {

	uuid, err := ws.socketBroker.NewSession()
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/session/%s", uuid.String()), http.StatusTemporaryRedirect)
}

func (ws WebServer) TTYHandler(w http.ResponseWriter, r *http.Request) {

	urlid := chi.URLParam(r, "uuid")
	if urlid == "" {
		ws.renderer.RenderBadRequest(w)
		return
	}

	id, err := uuid.Parse(urlid)
	if err != nil {
		ws.renderer.RenderBadRequest(w)
		return
	}

	if !ws.socketBroker.SessionExists(id) {
		ws.renderer.RenderBadRequest(w)
		return
	}

	err = ws.renderer.templates.ExecuteTemplate(w, "tty.html", nil)
	if err != nil {
		ws.renderer.RenderInternalError(w)
		log.Printf("DEBUG: failed to render tty template: %s\n", err.Error())
	}
}

func (ws WebServer) WebSocketHandler(w http.ResponseWriter, r *http.Request) {

	urlid := chi.URLParam(r, "uuid")
	if urlid == "" {
		ws.renderer.RenderBadRequest(w)
		return
	}

	id, err := uuid.Parse(urlid)
	if err != nil {
		ws.renderer.RenderBadRequest(w)
		return
	}

	if !ws.socketBroker.SessionExists(id) {
		ws.renderer.RenderBadRequest(w)
		return
	}

	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("DEBUG: err upgrading: %s\n", err.Error())
		ws.renderer.RenderInternalError(w)
		return
	}

	if err := ws.socketBroker.AddWebSocket(id, c); err != nil {
		log.Printf("DEBUG: err adding websocket: %s\n", err.Error())
		ws.renderer.RenderInternalError(w)
		return
	}
}
