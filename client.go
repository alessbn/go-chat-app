package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client and will hold a reference
	// to the websocket that will allow us to communicate with the client.
	socket *websocket.Conn
	// send is a buffered channel through which received messages are queued
	// ready to be forwarded to the user's browser.
	send chan *message
	// room  will keep a reference to the room that the client is chatting in
	// in order to forward messages to everyone else in the room.
	room *room
	// userData holds information about the user,
	// this user data comes from the client cookie.
	userData map[string]interface{}
}

// read method allows our client to read from the socket via ReadMessage method
// continually sending any received messages to the forward channel on the room type.
func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData["name"].(string)
		msg.AvatarURL, _ = c.room.avatar.GetAvatarURL(c)
		c.room.forward <- msg
	}
}

// write method continually accepts messages from the send channel writing everyting out of the socket
// via WriteMessage method.
func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}
