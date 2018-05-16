package piPlayer

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageSize  = 512
	readBufferSize  = 1024
	writeBufferSize = 1024
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
}

// connectionWS has the websocket connection to the control page
type connectionWS struct {
	conn   *websocket.Conn
	send   chan resMessage
	active bool
}

// HandleWebsocket handles websocket connections
func (c *connectionWS) HandlerWebsocket(w http.ResponseWriter, r *http.Request) {
	var err error
	c.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error trying to upgrade to websocket connection:", err)
		c.active = false
		return
	}

	c.active = true
	go c.write()
	go c.read()
}

// write sends data to the websocket
func (c *connectionWS) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// this loop keeps running as long as the channel is open.
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteJSON(msg)
			if err != nil {
				log.Println("Error trying to write JSON to the socket: ", err)
				// this probably means that the connection is broken,
				// so close the channel and break out of the loop.
				close(c.send)
				c.active = false
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Error trying to send ping message")
				// not sure what else needs to be done here.
				return
			}
		}
	}
}

// read reads the messages from the socket
func (c *connectionWS) read() {
	defer c.conn.Close()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg reqMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error, websocket closed: ", err)
			} else {
				log.Println("Error trying to read the JSON from the socket: ", err)
			}
			close(c.send)
			c.active = false
			break
		}

		// ignore socket messages for now.
		// TODO: handle socket messages.
		log.Println("socket message received: ", msg)
	}
}