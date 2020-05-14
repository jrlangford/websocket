package websocket

import (
	"encoding/json"
	"io"
)

// A JSONData stores partially decoded json data that includes the 'channel' key at its top level.
type JSONData struct {
	Channel string          `json:"channel"`
	Payload json.RawMessage `json:"payload"`
}

// A JSONChan is a channel of json.RawMessage
type JSONChan chan json.RawMessage

// A JSONRouter stores a map of JSONChan so each JSONChan can be accessed by its ID.
type JSONRouter struct {
	channels map[string]JSONChan
}

// NewJSONRouter creates an initialized JSONRouter and returns its pointer.
// Each JSONRouter contains a 'default' channel where messages to unknown routes are pushed.
func NewJSONRouter(channels []string, chLen int) (r *JSONRouter) {
	r = &JSONRouter{
		channels: make(map[string]JSONChan),
	}
	for _, s := range channels {
		r.channels[s] = make(chan json.RawMessage, chLen)
	}
	r.channels["default"] = make(chan json.RawMessage, chLen)
	return
}

// Get returns a JSONChan by its ID, if it is not found the 'default' JSONChan is returned.
func (j *JSONRouter) Get(id string) JSONChan {
	if v, ok := j.channels[id]; ok {
		return v
	}
	return j.channels["default"]
}

// Route reads data from the reader and pushes it to its corresponding channel
func (j *JSONRouter) Route(messageType int, r io.Reader) (err error) {
	jdata := &JSONData{}

	err = json.NewDecoder(r).Decode(jdata)
	if err != nil {
		return
	}

	j.Get(jdata.Channel) <- jdata.Payload
	return
}
