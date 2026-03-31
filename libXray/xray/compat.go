package xray

import (
	"os"
	"sync"
	"time"

	"github.com/xtls/xray-core/common/platform"
)

var (
	startTime  time.Time
	currentURL string
	mu         sync.Mutex
)

// GetDatDir returns the current asset directory
func GetDatDir() string {
	return os.Getenv(platform.AssetLocation)
}

// SetCurrentURL sets the currently used URL
func SetCurrentURL(url string) {
	mu.Lock()
	currentURL = url
	mu.Unlock()
}

// GetUsing returns the currently used URL
func GetUsing() string {
	mu.Lock()
	defer mu.Unlock()
	return currentURL
}

// Duration returns how long the VPN has been running in seconds
func Duration() int64 {
	if !GetXrayState() {
		return 0
	}
	mu.Lock()
	defer mu.Unlock()
	return int64(time.Since(startTime).Seconds())
}

// recordStartTime records the VPN start timestamp
func recordStartTime() {
	mu.Lock()
	startTime = time.Now()
	mu.Unlock()
}
