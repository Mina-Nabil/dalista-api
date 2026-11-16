package resources

import (
	"encoding/json"
	"time"

	notificationModels "dalista/dalista-be/app/repository/models/notificationModels"
)

// DeviceTokenResource is a registered push device.
type DeviceTokenResource struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
}

// DeviceTokenFromModel maps a device token to its resource.
func DeviceTokenFromModel(d *notificationModels.DeviceToken) DeviceTokenResource {
	return DeviceTokenResource{ID: d.ID.String(), Platform: string(d.Platform)}
}

// NotificationResource is one inbox notification.
type NotificationResource struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data"`
	CreatedAt string            `json:"created_at"`
}

// NotificationFromModel maps a notification to its resource.
func NotificationFromModel(n *notificationModels.Notification) NotificationResource {
	data := map[string]string{}
	if n.Data != "" {
		_ = json.Unmarshal([]byte(n.Data), &data)
	}
	return NotificationResource{
		ID:        n.ID.String(),
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Data:      data,
		CreatedAt: n.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// NotificationsFromModels maps a slice of notifications.
func NotificationsFromModels(ns []notificationModels.Notification) []NotificationResource {
	out := make([]NotificationResource, 0, len(ns))
	for i := range ns {
		out = append(out, NotificationFromModel(&ns[i]))
	}
	return out
}
