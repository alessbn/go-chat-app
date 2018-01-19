package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"trace"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan []byte
	// join us a channel for clients wishing to join the room and
	// allows us to safely add clients from the clients map.
	join chan *client
	// leave is a channel for clients wishing to leave the room and
	// allows us to safely remove clients from the clients map.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
	// tracer will receive trace information of activity
	// in the room
	tracer trace.Tracer
}

// newRoom makes a new room.
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		// only one block of case code will run at a time,
		// this is how we are able to synchronize to ensure that our
		// r.clients map is only ever modified by one thing at a time.
		select {
		case client := <-r.join:
			// if join channel receive a message
			// r.clients map will be update to keep a reference of the client that has joined the room.
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			// if leave channel receive a message
			// will delete the client type from the map and close its send channel.
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Messages received: ", string(msg))
			// if forward channel receive a message
			// will iterate over all the clients and adds the message to each client's send channel
			// then write method of client type will pick up and send it down the socket to the browser.
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// in order to use web sockets, we must upgrade the HTTP connection using the websocket.Upgrader type
// which is reusable so we need only create one.
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// ServeHTTP method means a room can now act as a hanlder.
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// when a request comes in via the ServeHTTP method, we get a socket by
	// calling the upgrader.Upgrade method
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	// then create our client ans pass it into the join channel for the current room
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	// we call read method in the main thread, wich will block operations
	// until it's time to close it.
	client.read()
}
