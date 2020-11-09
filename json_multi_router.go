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
type JSONChan struct {
	Data chan json.RawMessage
	Flow bool
}

// NewJSONChan initialises a JSONChan and returns a pointer to it
func NewJSONChan(flow bool, chLen int) *JSONChan {
	return &JSONChan{
		Data: make(chan json.RawMessage, chLen),
		Flow: flow,
	}
}

// A JSONRouter stores a map of JSONChan so each JSONChan can be accessed by its ID.
type JSONRouter struct {
	channels map[string]*JSONChan
}

// NewJSONRouter creates an initialized JSONRouter and returns its pointer.
func NewJSONRouter(channels []string, chLen int) (r *JSONRouter) {
	r = &JSONRouter{
		channels: make(map[string]*JSONChan),
	}
	for _, s := range channels {
		r.channels[s] = NewJSONChan(false, chLen)
	}
	return
}

// Get returns a JSONChan by its ID, if it is not found a JSONChan with Ignore:true is returned to allow chaining without critical failure.
func (j *JSONRouter) Get(id string) *JSONChan {
	v, ok := j.channels[id]
	if ok {
		return v
	}
	return &JSONChan{Flow: false}
}

// Route reads data from the reader and pushes it to its corresponding channel
func (j *JSONRouter) Route(messageType int, r io.Reader) (err error) {
	jdata := &JSONData{}

	err = json.NewDecoder(r).Decode(jdata)
	if err != nil {
		return
	}

	c := j.Get(jdata.Channel)
	if c.Flow {
		c.Data <- jdata.Payload
	}
	return
}
