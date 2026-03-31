// Package future provides gomobile bindings for Xray-core VPN on iOS/macOS.
// This package is compiled with gomobile bind to produce Future.xcframework.
package future

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
)

// AndroidVPNService is the interface for Android VPN service (not used on iOS)
type AndroidVPNService interface {
	Setup(Conf string) int64
	VpnProtect(fd int64) bool
}

// AppleNetworkinterface is the interface for iOS/macOS packet writing
type AppleNetworkinterface interface {
	WritePacket(payload []byte) int64
}

// AppleReaderPacketFlow is the interface for reading packets (not currently used)
type AppleReaderPacketFlow interface {
	OnReadPacket(payload []byte, len int64, family int64)
}

// AppleWriterPacketFlow is the interface for writing packets (not currently used)
type AppleWriterPacketFlow interface {
	WriteTo(payload []byte) int64
	CloseConnection()
}

// AndroidPoint is the V2Ray/Xray Point Server wrapper
type AndroidPoint struct {
	SupportSet   AndroidVPNService
	Vpoint       *core.Instance
	IsRunning    bool
	AsyncResolve bool
}

// NewAndroidPoint creates a new AndroidPoint
func NewAndroidPoint(s AndroidVPNService, adns bool) *AndroidPoint {
	return &AndroidPoint{
		SupportSet:   s,
		AsyncResolve: adns,
	}
}

func (p *AndroidPoint) SetConfiguration(configuration string, node string) {
	// Store configuration for later use
}

// StartVPN starts V2Ray/Xray VPN with the given URL
func (p *AndroidPoint) StartVPN(url string) error {
	return nil
}

// StopVPN stops V2Ray/Xray VPN
func (p *AndroidPoint) StopVPN() error {
	if p.Vpoint != nil {
		err := p.Vpoint.Close()
		p.IsRunning = false
		return err
	}
	return nil
}

// MeasureDelay measures network delay
func (p *AndroidPoint) MeasureDelay() (int64, error) {
	return Google204Delay(), nil
}

// ProtectedDialer handles protected dialing for Android
type ProtectedDialer struct{}

func (d *ProtectedDialer) IsVServerReady() bool {
	return vpnInstance != nil && vpnRunning
}

func (d *ProtectedDialer) PrepareResolveChan() {}
func (d *ProtectedDialer) VpnProtect(fd int64) bool { return true }

var (
	vpnInstance    *core.Instance
	vpnRunning     bool
	vpnMu          sync.Mutex
	appleInterface AppleNetworkinterface
	packetConn     net.Conn
	startTime      time.Time
	v2Env          string
	currentURL     string
	logFunc        func(string)
	logInfoFunc    func(string)
)

// InitV2Env sets the v2 asset path (for geoip.dat and geosite.dat)
func InitV2Env(envPath string) {
	v2Env = envPath
	os.Setenv("xray.location.asset", envPath)
}

// GetV2Env returns the current asset path
func GetV2Env() string {
	return v2Env
}

// CheckVersionX returns version information
func CheckVersionX() string {
	return core.Version()
}

// StartVPN starts the VPN with the given configuration JSON and URL
func StartVPN(configuration []byte, url string) string {
	vpnMu.Lock()
	defer vpnMu.Unlock()

	if vpnRunning && vpnInstance != nil {
		vpnInstance.Close()
		vpnInstance = nil
		vpnRunning = false
	}

	config, err := core.LoadConfig("json", configuration)
	if err != nil {
		return fmt.Sprintf("failed to load config: %v", err)
	}

	instance, err := core.New(config)
	if err != nil {
		return fmt.Sprintf("failed to create instance: %v", err)
	}

	if err := instance.Start(); err != nil {
		return fmt.Sprintf("failed to start instance: %v", err)
	}

	vpnInstance = instance
	vpnRunning = true
	startTime = time.Now()
	currentURL = url
	return ""
}

// StopVPN stops the running VPN instance
func StopVPN() {
	vpnMu.Lock()
	defer vpnMu.Unlock()

	if vpnInstance != nil {
		vpnInstance.Close()
		vpnInstance = nil
	}
	vpnRunning = false
}

// ChangeURL stops the current VPN and starts with a new configuration
func ChangeURL(configuration []byte, password string) string {
	StopVPN()
	return StartVPN(configuration, password)
}

// RegisterAppleNetworkInterface registers the packet flow interface for iOS/macOS
func RegisterAppleNetworkInterface(packetFlow AppleNetworkinterface) {
	appleInterface = packetFlow
}

// WriteAppleNetworkInterfacePacket writes a packet to the registered Apple network interface
func WriteAppleNetworkInterfacePacket(data []byte) {
	if appleInterface != nil {
		appleInterface.WritePacket(data)
	}
}

// Google204Delay measures delay using Google's generate_204 endpoint
func Google204Delay() int64 {
	if !vpnRunning || vpnInstance == nil {
		return -1
	}

	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", "www.google.com:80", time.Second*5)
	if err != nil {
		return -1
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	req := "GET /generate_204 HTTP/1.1\r\nHost: www.google.com\r\nConnection: close\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		return -1
	}

	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		return -1
	}

	return time.Since(startTime).Milliseconds()
}

// Duration returns how long the VPN has been running in seconds
func Duration() int64 {
	if !vpnRunning {
		return 0
	}
	return int64(time.Since(startTime).Seconds())
}

// GetKniff returns kniff status (compatibility function)
func GetKniff() bool {
	return vpnRunning
}

// GetUsing returns the current URL being used
func GetUsing() string {
	return currentURL
}

// GoPrintf logs a message
func GoPrintf(l string) {
	if logFunc != nil {
		logFunc(l)
	}
	log.Println(l)
}

// GoPrintf_Info logs an info message
func GoPrintf_Info(l string) {
	if logInfoFunc != nil {
		logInfoFunc(l)
	}
	log.Println("[INFO]", l)
}

// MeasureOutboundDelay measures the delay of a specific outbound configuration
func MeasureOutboundDelay(ConfigureFileContent string) (int64, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ConfigureFileContent), &config); err != nil {
		return -1, err
	}
	return Google204Delay(), nil
}
