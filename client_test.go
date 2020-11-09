package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewWebSocket(t *testing.T) {
	jr := NewJSONRouter([]string{"a", "b", "c", "d"}, 20)
	ws := New(jr)

	assert.NotNil(t, ws.Router)
}

func NewEchoHandler(tb testing.TB, router WSRouter, exit *bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws := New(router)
		err := ws.Connect(w, r)
		if err != nil {
			tb.Fatal("Echo Handler initialization failed: " + err.Error())
		}
		defer ws.Stop()

		loop := true
		for loop {
			select {
			case data := <-ws.Router.Get("dp").Data:
				//tb.Logf("Received %s", data)
				err := ws.WriteTextMessage([]byte(fmt.Sprintf("%s", data)))
				if err != nil {
					tb.Logf("TX Err: %+v", err.Error())
					loop = false
					break
				}
			case err := <-ws.Errors:
				tb.Logf("RX Err: %+v", err)
				loop = false
				break
			}
		}
		if exit != nil {
			*exit = true
		}
	}
}

func TestConnect(t *testing.T) {

	echoServerExited := false

	router := NewJSONRouter([]string{"dp"}, 20)
	router.Get("dp").Flow = true

	echo := NewEchoHandler(t, router, &echoServerExited)

	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	wsClient, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsClient.Close()

	t.Log("Client Ready")

	data := struct {
		Channel string `json:"channel"`
		Payload string `json:"payload"`
	}{
		Channel: "dp",
		Payload: "regular string",
	}

	messageJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := wsClient.WriteMessage(websocket.TextMessage, []byte(messageJSON)); err != nil {
			t.Fatalf("%v", err)
		}
		t.Log("Attempting Read")
		_, p, err := wsClient.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}

		assert.Equal(t, "\"regular string\"", string(p))
		t.Logf("Round %d successful", i)
	}

	wsClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client closed connection"))
	time.Sleep(1 * time.Second)
	wsClient.Close()

	assert.True(t, echoServerExited, "Echo server did not shut down correctly")

}

func BenchmarkMultipleRequests(b *testing.B) {

	router := NewJSONRouter([]string{"dp"}, 20)

	echo := NewEchoHandler(b, router, nil)

	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	wsClient, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		b.Fatalf("%v", err)
	}
	defer wsClient.Close()

	data := struct {
		Channel string `json:"channel"`
		Payload string `json:"payload"`
	}{
		Channel: "dp",
		Payload: "regular string",
	}

	messageJSON, err := json.Marshal(data)
	if err != nil {
		b.Fatal(err.Error())
	}

	for n := 0; n < b.N; n++ {
		err := wsClient.WriteMessage(websocket.TextMessage, []byte(messageJSON))
		if err != nil {
			b.Fatalf("%v", err)
			break
		}
		_, _, err = wsClient.ReadMessage()
		if err != nil {
			b.Fatalf("%v", err)
			break
		}
	}

	//wsClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client closed connection"))
	//time.Sleep(1 * time.Second)
	//wsClient.Close()
}
