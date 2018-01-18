package main

import (
	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client and will hold a reference
	// to the websocket that will allow us to communicate with the client.
	socket *websocket.Conn
	// send is a buffered channel through which received messages are queued
	// ready to be forwarded to the user's browser.
	send chan []byte
	// room  will keep a reference to the room that the client is chatting in
	// in order to forward messages to everyone else in the room.
	room *room
}

// read method allows our client to read from the socket via ReadMessage method
// continually sending any received messages to the forward channel on the room type.
func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

// write method continually accepts messages from the send channel writing everyting out of the socket
// via WriteMessage method.
func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
