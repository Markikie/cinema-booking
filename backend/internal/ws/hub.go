package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type SeatEvent struct {
	Type       string `json:"type"` // "SEAT_LOCKED" | "SEAT_RELEASED" | "SEAT_BOOKED"
	ShowtimeID string `json:"showtime_id"`
	SeatID     string `json:"seat_id"`
	Status     string `json:"status"`
}

type client struct {
	conn       *websocket.Conn
	showtimeID string
	send       chan []byte // buffered channel กันไม่ให้ hub รอ client ที่ช้าจนบล็อกคนอื่น
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]bool),
	}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

func (h *Hub) Broadcast(event SeatEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("failed to marshal seat event:", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		if c.showtimeID != event.ShowtimeID {
			continue // ข้าม client ที่ดูรอบฉายอื่นอยู่
		}
		select {
		case c.send <- payload:
			// ส่งสำเร็จ
		default:
			// channel เต็ม (client อ่านไม่ทัน/ค้าง) - ไม่รอ ไม่งั้นจะบล็อก broadcast
			// ทั้งหมดเพราะ client ตัวเดียวที่มีปัญหา ปิด connection นั้นทิ้งไปเลย
			log.Println("client send buffer full, dropping connection")
			go h.unregister(c)
		}
	}
}
