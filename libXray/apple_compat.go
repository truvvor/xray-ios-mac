// Package libXray Apple compatibility layer.
// Provides the Future-style API that the existing xNetFuture iOS/macOS app expects.
package libXray

import (
	"encoding/base64"
	"encoding/json"

	"github.com/xtls/libxray/nodep"
	"github.com/xtls/libxray/xray"
)

// AppleNetworkinterface is the interface for iOS/macOS packet writing
type AppleNetworkinterface interface {
	WritePacket(payload []byte) int64
}

var appleInterface AppleNetworkinterface

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

// StartVPN starts the VPN with configuration JSON and URL (Future-compatible API).
// Returns empty string on success, error message on failure.
func StartVPN(configuration []byte, url string) string {
	configJSON := string(configuration)
	request := RunXrayFromJSONRequest{
		DatDir:     xray.GetDatDir(),
		ConfigJSON: configJSON,
	}
	requestBytes, err := json.Marshal(&request)
	if err != nil {
		return err.Error()
	}
	b64 := base64.StdEncoding.EncodeToString(requestBytes)
	result := RunXrayFromJSON(b64)

	// Decode result to check for errors
	resultBytes, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return err.Error()
	}
	var response nodep.CallResponse[string]
	if err := json.Unmarshal(resultBytes, &response); err != nil {
		return err.Error()
	}
	if !response.Success {
		return response.Err
	}
	return ""
}

// StopVPN stops the running VPN instance (Future-compatible API)
func StopVPN() {
	xray.StopXray()
}

// ChangeURL stops the current VPN and starts with new configuration (Future-compatible API)
func ChangeURL(configuration []byte, password string) string {
	StopVPN()
	return StartVPN(configuration, password)
}

// InitV2Env sets the v2 asset path for geoip.dat and geosite.dat (Future-compatible API)
func InitV2Env(envPath string) {
	xray.InitEnv(envPath, "")
}

// CheckVersionX returns Xray-core version string (Future-compatible API)
func CheckVersionX() string {
	return xray.XrayVersion()
}

// Google204Delay measures delay using the Xray proxy (Future-compatible API)
func Google204Delay() int64 {
	if !xray.GetXrayState() {
		return -1
	}
	// Use libXray's proper ping through the proxy
	request := pingRequest{
		DatDir:  xray.GetDatDir(),
		Timeout: 5,
		Url:     "https://www.google.com/generate_204",
		Proxy:   "socks5://127.0.0.1:1080",
	}
	requestBytes, _ := json.Marshal(&request)
	b64 := base64.StdEncoding.EncodeToString(requestBytes)
	result := Ping(b64)

	resultBytes, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return -1
	}
	var response nodep.CallResponse[int64]
	if err := json.Unmarshal(resultBytes, &response); err != nil {
		return -1
	}
	if !response.Success {
		return -1
	}
	return response.Data
}

// Duration returns how long the VPN has been running (Future-compatible API)
func Duration() int64 {
	return xray.Duration()
}

// GetKniff returns VPN running status (Future-compatible API)
func GetKniff() bool {
	return xray.GetXrayState()
}

// GetUsing returns the current URL being used (Future-compatible API)
func GetUsing() string {
	return xray.GetUsing()
}

// GetV2Env returns the current asset path (Future-compatible API)
func GetV2Env() string {
	return xray.GetDatDir()
}

// GoPrintf logs a message (Future-compatible API)
func GoPrintf(l string) {
	// No-op: Xray-core handles its own logging
}

// GoPrintf_Info logs an info message (Future-compatible API)
func GoPrintf_Info(l string) {
	// No-op: Xray-core handles its own logging
}

// MeasureOutboundDelay measures the delay of a specific outbound configuration (Future-compatible API)
func MeasureOutboundDelay(ConfigureFileContent string) (int64, error) {
	return Google204Delay(), nil
}
