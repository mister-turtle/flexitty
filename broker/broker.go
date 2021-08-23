package broker

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mister-turtle/flexitty/tty"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	WS  []*websocket.Conn
	TTY *tty.TTY
	mu  sync.Mutex
}

func (s *Session) Broadcast(msgtype int, data []byte) {

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, socket := range s.WS {
		err := socket.WriteMessage(msgtype, data)
		if err != nil {
			log.Printf("ERR: web socket WriteMessage failed: %s\n", err.Error())
		}
	}
}

type Broker struct {
	sessions map[uuid.UUID]*Session
	mu       sync.Mutex
}

func (b *Broker) NewSession() (uuid.UUID, error) {

	id := uuid.New()
	newTTY, err := tty.New("bash", []string{})
	if err != nil {
		return id, err
	}

	session := &Session{
		TTY: newTTY,
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.sessions[id] = session

	// Start a go routine to read from the TTY and send to any available web sockets
	go func() {
		for {
			var data []byte
			data, err := session.TTY.Read()
			if err != nil {
				log.Printf("DEBUG: ERR: failed to read from TTY\n")
				continue
			}
			session.Broadcast(websocket.TextMessage, data)
		}
	}()
	return id, nil
}

func (b *Broker) AddWebSocket(id uuid.UUID, c *websocket.Conn) error {

	if session, ok := b.sessions[id]; ok {
		session.mu.Lock()
		defer session.mu.Unlock()
		session.WS = append(session.WS, c)

		// Start a go routine handling input from the web socket
		go func() {
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("DEBUG: failed to ReadMessage(): %s\n", err.Error())
					break
				}
				session.TTY.Write(message)
			}
		}()
		return nil
	}
	return ErrSessionNotFound
}

func (b *Broker) SessionExists(id uuid.UUID) bool {
	_, ok := b.sessions[id]
	return ok
}

func New() *Broker {
	return &Broker{
		sessions: map[uuid.UUID]*Session{},
		mu:       sync.Mutex{},
	}
}
