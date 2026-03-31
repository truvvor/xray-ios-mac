package encryption

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"sync"
	"time"
)

// HeartbeatConn wraps a connection and sends fake TLS Application Data records
// during idle periods to prevent DPI from detecting sleeping tunnels.
type HeartbeatConn struct {
	net.Conn
	minIntervalMs int
	maxIntervalMs int

	mu       sync.Mutex
	lastIO   time.Time
	stopCh   chan struct{}
	stopped  bool
}

func NewHeartbeatConn(conn net.Conn, minIntervalMs, maxIntervalMs int) *HeartbeatConn {
	h := &HeartbeatConn{
		Conn:          conn,
		minIntervalMs: minIntervalMs,
		maxIntervalMs: maxIntervalMs,
		lastIO:        time.Now(),
		stopCh:        make(chan struct{}),
	}
	go h.heartbeatLoop()
	return h
}

func (h *HeartbeatConn) heartbeatLoop() {
	for {
		interval := heartbeatRandInt(h.minIntervalMs, h.maxIntervalMs)
		select {
		case <-h.stopCh:
			return
		case <-time.After(time.Duration(interval) * time.Millisecond):
		}

		h.mu.Lock()
		idle := time.Since(h.lastIO)
		h.mu.Unlock()

		// Only send heartbeat if idle > 2 seconds
		if idle < 2*time.Second {
			continue
		}

		// Generate fake TLS Application Data record
		payloadLen := heartbeatRandInt(16, 128)
		record := make([]byte, 5+payloadLen)
		record[0] = 0x17       // TLS Application Data
		record[1] = 0x03       // TLS 1.2 major version
		record[2] = 0x03       // TLS 1.2 minor version
		record[3] = byte(payloadLen >> 8)
		record[4] = byte(payloadLen)
		rand.Read(record[5:])

		h.Conn.Write(record) // Best effort, ignore error
	}
}

func (h *HeartbeatConn) Write(b []byte) (int, error) {
	h.mu.Lock()
	h.lastIO = time.Now()
	h.mu.Unlock()
	return h.Conn.Write(b)
}

func (h *HeartbeatConn) Read(b []byte) (int, error) {
	n, err := h.Conn.Read(b)
	h.mu.Lock()
	h.lastIO = time.Now()
	h.mu.Unlock()
	return n, err
}

func (h *HeartbeatConn) Close() error {
	h.mu.Lock()
	if !h.stopped {
		h.stopped = true
		close(h.stopCh)
	}
	h.mu.Unlock()
	return h.Conn.Close()
}

func heartbeatRandInt(min, max int) int {
	if min >= max {
		return min
	}
	var buf [4]byte
	rand.Read(buf[:])
	n := int(binary.LittleEndian.Uint32(buf[:])) % (max - min + 1)
	return min + n
}
