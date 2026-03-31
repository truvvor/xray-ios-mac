package encryption

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"sync"
	"time"
)

// ScatterConn wraps a connection and splits writes across multiple TCP segments
// to break the assumption that TLS records align with TCP segment boundaries.
type ScatterConn struct {
	net.Conn
	minChunk    int
	maxChunk    int
	maxScatter  int
	maxJitterMs int

	mu         sync.Mutex
	writeCount int
}

func NewScatterConn(conn net.Conn, minChunk, maxChunk, maxScatter, maxJitterMs int) *ScatterConn {
	return &ScatterConn{
		Conn:        conn,
		minChunk:    minChunk,
		maxChunk:    maxChunk,
		maxScatter:  maxScatter,
		maxJitterMs: maxJitterMs,
	}
}

func (s *ScatterConn) Write(b []byte) (int, error) {
	s.mu.Lock()
	s.writeCount++
	count := s.writeCount
	s.mu.Unlock()

	// After maxScatter writes, pass through without scattering
	if s.maxScatter > 0 && count > s.maxScatter {
		return s.Conn.Write(b)
	}

	// Don't bother scattering small writes
	if len(b) < s.minChunk*2 {
		return s.Conn.Write(b)
	}

	total := 0
	for len(b) > 0 {
		chunkSize := scatterRandInt(s.minChunk, s.maxChunk)
		if chunkSize > len(b) {
			chunkSize = len(b)
		}
		n, err := s.Conn.Write(b[:chunkSize])
		total += n
		if err != nil {
			return total, err
		}
		b = b[chunkSize:]
		if len(b) > 0 && s.maxJitterMs > 0 {
			jitter := scatterRandInt(0, s.maxJitterMs)
			if jitter > 0 {
				time.Sleep(time.Duration(jitter) * time.Millisecond)
			}
		}
	}
	return total, nil
}

func scatterRandInt(min, max int) int {
	if min >= max {
		return min
	}
	var buf [4]byte
	rand.Read(buf[:])
	n := int(binary.LittleEndian.Uint32(buf[:])) % (max - min + 1)
	return min + n
}
