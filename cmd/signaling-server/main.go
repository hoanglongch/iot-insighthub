package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// SignalingMessage defines the JSON structure for signaling messages.
type SignalingMessage struct {
	Type    string          `json:"type"`    // "join", "offer", "answer", "candidate"
	From    string          `json:"from"`    // sender's client ID
	To      string          `json:"to"`      // recipient's client ID
	Payload json.RawMessage `json:"payload"` // SDP details or ICE candidate, etc.
}

// Client represents a connected WebSocket client.
type Client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
	hub  *Hub
}

// readPump handles incoming messages from the client.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read error from client %s: %v", c.id, err)
			break
		}
		// Log incoming raw message for debugging.
		log.Printf("Received message from %s: %s", c.id, message)
		// Forward the message to the hub for routing.
		c.hub.broadcast <- message
	}
}

// writePump sends outgoing messages to the client.
func (c *Client) writePump() {
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("write error to client %s: %v", c.id, err)
			break
		}
	}
	c.conn.Close()
}

// Hub maintains the set of active clients and routes messages.
type Hub struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			log.Printf("Client %s registered", client.id)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
				log.Printf("Client %s unregistered", client.id)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			// Parse the signaling message.
			var msg SignalingMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}
			// Route the message to the intended recipient.
			h.mu.RLock()
			target, exists := h.clients[msg.To]
			h.mu.RUnlock()
			if exists {
				target.send <- message
				log.Printf("Forwarded message from %s to %s", msg.From, msg.To)
			} else {
				log.Printf("Target client %s not found", msg.To)
			}
		}
	}
}

// Upgrader for WebSocket connection.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for this example. In production, limit origins.
		return true
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Expect a query parameter "id" for the client identifier.
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		http.Error(w, "missing client id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &Client{
		id:   clientID,
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
	}
	hub.register <- client

	// Start read and write pumps.
	go client.readPump()
	go client.writePump()
}

func main() {
	addr := flag.String("addr", ":8081", "http service address")
	flag.Parse()
	hub := newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Printf("Signaling server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
