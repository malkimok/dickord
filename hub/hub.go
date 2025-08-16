package hub

type BroadcastMessage struct {
	author  *ClientConnection
	message []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*ClientConnection]bool

	// Inbound messages from the clients.
	broadcast chan BroadcastMessage

	// Register requests from the clients.
	register chan *ClientConnection

	// Unregister requests from clients.
	unregister chan *ClientConnection
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan BroadcastMessage),
		register:   make(chan *ClientConnection),
		unregister: make(chan *ClientConnection),
		clients:    make(map[*ClientConnection]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if client == message.author {
					continue
				}
				select {
				case client.send <- message.message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
