package websocket

import (
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

// Signal is an empty struct used to indicate events by sending it through a Signal channel.
type Signal struct{}

// WSRouter is the interface that defines Websocket router signatures.
// Get returns a JSONChan by its ID.
// Route places the Reader's data into its corresponding JSONChan.
type WSRouter interface {
	Get(id string) JSONChan
	Route(messageType int, r io.Reader) error
}

// A WebSocket stores its Router, Errors, and underlying websockets connection.
type WebSocket struct {
	Router WSRouter
	Errors chan error
	//stopchan chan Signal
	Conn *websocket.Conn
}

// Stop shuts down the resources used by the WebSocket.
func (ws *WebSocket) Stop() {
	ws.Conn.Close()
}

// WriteTextMessage sends a []byte to the connection as a websocket.TextMessage.
func (ws *WebSocket) WriteTextMessage(data []byte) error {
	return ws.Conn.WriteMessage(websocket.TextMessage, data)
}

// New creates a new WebSocket and returns its pointer.
func New(r WSRouter) *WebSocket {
	return &WebSocket{
		//TX: NewCommChan(),
		Errors: make(chan error, 20),
		Router: r,
		//RX:     NewCommChan(),
	}
}

// Connect upgrades the conenction and launches a goroutine that listens for incoming data and sends it to the router
func (ws *WebSocket) Connect(w http.ResponseWriter, r *http.Request) (err error) {

	upgrader := websocket.Upgrader{}
	// TODO: Set an Origin policy, currently all are allowed
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws.Conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		err = errors.New("upgrade: " + err.Error())
		return
	}

	// Launches a reader goroutine. It can only be unblocked by closing the connection or the underlying reader in conn.ReadMessage()
	go func() {
		for {
			messageType, r, err := ws.Conn.NextReader()
			if err != nil {
				ws.Conn.Close()
				ws.Errors <- err
				break
			}
			err = ws.Router.Route(messageType, r)
			if err != nil {
				ws.Errors <- err
				continue
			}
		}
	}()
	return
}
