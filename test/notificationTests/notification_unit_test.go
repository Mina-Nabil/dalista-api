package notificationTests

import (
	"testing"
	"time"

	"dalista/dalista-be/api/resources"
	notificationModels "dalista/dalista-be/app/repository/models/notificationModels"

	"github.com/google/uuid"
)

func TestDeviceTokenResourceMapping(t *testing.T) {
	d := &notificationModels.DeviceToken{ID: uuid.New(), Platform: notificationModels.PlatformAndroid}
	r := resources.DeviceTokenFromModel(d)
	if r.ID != d.ID.String() || r.Platform != "android" {
		t.Fatalf("device token mapping wrong: %+v", r)
	}
}

func TestNotificationResourceParsesData(t *testing.T) {
	n := &notificationModels.Notification{
		ID:        uuid.New(),
		Type:      "list_shared",
		Title:     "A list was shared with you",
		Body:      "You now have access to a shared list.",
		Data:      `{"list_id":"abc","role":"editor"}`,
		CreatedAt: time.Now(),
	}
	r := resources.NotificationFromModel(n)
	if r.Type != "list_shared" || r.Title == "" {
		t.Fatalf("notification mapping wrong: %+v", r)
	}
	if r.Data["list_id"] != "abc" || r.Data["role"] != "editor" {
		t.Fatalf("data parse wrong: %+v", r.Data)
	}
}

func TestNotificationResourceEmptyData(t *testing.T) {
	n := &notificationModels.Notification{ID: uuid.New(), Type: "x", Title: "t", Body: "b", Data: ""}
	r := resources.NotificationFromModel(n)
	if r.Data == nil || len(r.Data) != 0 {
		t.Fatalf("empty data should map to empty map, got %+v", r.Data)
	}
}
