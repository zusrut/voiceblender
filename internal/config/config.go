package config

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

type Config struct {
	InstanceID   string
	SIPBindIP    string
	SIPListenIP  string
	SIPPort      string
	SIPHost      string
	HTTPAddr     string
	ICEServers   []string
	RecordingDir string
	LogLevel     string
	EnablePprof       bool
	WebhookURL        string
	ElevenLabsAPIKey  string
	VAPIAPIKey        string
	S3Bucket   string
	S3Region   string
	S3Endpoint string
	S3Prefix   string
}

func Load() Config {
	return Config{
		InstanceID:   envOr("INSTANCE_ID", uuid.New().String()),
		SIPBindIP:    envOr("SIP_BIND_IP", "127.0.0.1"),
		SIPListenIP:  os.Getenv("SIP_LISTEN_IP"), // empty = same as SIP_BIND_IP
		SIPPort:      envOr("SIP_PORT", "5060"),
		SIPHost:      envOr("SIP_HOST", "voiceblender"),
		HTTPAddr:     envOr("HTTP_ADDR", ":8080"),
		ICEServers:   strings.Split(envOr("ICE_SERVERS", "stun:stun.l.google.com:19302"), ","),
		RecordingDir: envOr("RECORDING_DIR", "/tmp/recordings"),
		LogLevel:     envOr("LOG_LEVEL", "info"),
		EnablePprof:       os.Getenv("ENABLE_PPROF") == "true",
		WebhookURL:        os.Getenv("WEBHOOK_URL"),
		ElevenLabsAPIKey:  os.Getenv("ELEVENLABS_API_KEY"),
		VAPIAPIKey:        os.Getenv("VAPI_API_KEY"),
		S3Bucket:   os.Getenv("S3_BUCKET"),
		S3Region:   envOr("S3_REGION", "us-east-1"),
		S3Endpoint: os.Getenv("S3_ENDPOINT"),
		S3Prefix:   os.Getenv("S3_PREFIX"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
