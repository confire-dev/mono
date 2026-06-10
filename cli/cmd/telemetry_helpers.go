package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/confire-dev/confire/auth"
	"github.com/confire-dev/confire/transport"
)

// postCLIEvent fires a lightweight event to the Worker from a CLI command context
// (outside the daemon). No-op when not logged in. Always fire-and-forget.
func postCLIEvent(eventType string) {
	key := resolveAPIKey()
	if key == "" {
		return
	}
	deviceID, _ := auth.DeviceID()
	payload := telemetryPayload{
		EventID:    newEventID(),
		EventType:  eventType,
		CLIVersion: buildVersion,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, workerURLEnv()+"/v1/events", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("X-Confire-Device", deviceID)
	for k, v := range transport.SignedHeaders(key, deviceID, body) {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
