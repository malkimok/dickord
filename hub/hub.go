package hub

import (
	"log"
)

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

// func mixBytesFloat32(a, b []byte) []byte {
// 	fa := byteToFloat(a)
// 	fb := byteToFloat(b)
// 	n := min(len(a), len(b))
// 	out := make([]float32, n)
// 	for i := range n {
// 		v := fa[i] + fb[i]
// 		out[i] = max(-1, min(v, 1))
// 	}
// 	return floatToByte(out)
// }

// func (h *Hub) _run() {
// 	frame_buff := make(chan BroadcastMessage, 100)
// 	ticker := time.NewTicker(time.Millisecond * 5)

// 	flushToClients := func() {
// 		messages := make([]BroadcastMessage, 100)
// 	Penis:
// 		for {
// 			select {
// 			case m := <-frame_buff:
// 				messages = append(messages, m)
// 			default:
// 				break Penis
// 			}
// 		}

// 		for client := range h.clients {
// 			client_messages := make([][]byte, 100)
// 			max_m_length := 0
// 			for _, m := range messages {
// 				if client == m.author {
// 					// log.Println("debug: client skiped while broadcasting:", client.conn.RemoteAddr().String())
// 					continue
// 				} else {
// 					max_m_length = max(max_m_length, len(m.message))
// 					client_messages = append(client_messages, m.message)
// 				}
// 			}

// 			client_out := make([]byte, max_m_length)
// 			for _, m := range client_messages {
// 				mixBytesFloat32(client_out, m)
// 			}

// 			select {
// 			case client.send <- client_out:
// 			default:
// 				log.Printf("warning: closing client %v cause sending message to send channel was blocking", client.conn.RemoteAddr().String())
// 				close(client.send)
// 				delete(h.clients, client)
// 			}
// 		}

// 	}

// 	for {
// 		select {
// 		case client := <-h.register:
// 			h.clients[client] = true
// 			log.Println("info: new client registered:", client.conn.RemoteAddr().String())
// 		case client := <-h.unregister:
// 			if _, ok := h.clients[client]; ok {
// 				log.Println("info: client unregistered:", client.conn.RemoteAddr().String())
// 				delete(h.clients, client)
// 				close(client.send)
// 			}
// 		case message := <-h.broadcast:
// 			log.Printf("debug: client %v broadcasted a meessage\n", message.author.conn.RemoteAddr().String())
// 			select {
// 			case frame_buff <- message:
// 			default:
// 				log.Println("debug: flush due to Frame buff full")
// 				flushToClients()
// 			}
// 		case <-ticker.C:
// 			log.Println("debug: Server tick")
// 			flushToClients()
// 		}
// 	}
// }

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("info: new client registered:", client.conn.RemoteAddr().String())
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Println("info: client unregistered:", client.conn.RemoteAddr().String())
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				log.Printf("debug: client %v broadcasted a meessage\n", client.conn.RemoteAddr().String())
				if client == message.author {
					log.Println("debug: client skiped while broadcasting:", client.conn.RemoteAddr().String())
					continue
				}
				select {
				case client.send <- message.message:
				default:
					log.Printf("warning: closing client %v cause sending message to send channel was blocking", client.conn.RemoteAddr().String())
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
