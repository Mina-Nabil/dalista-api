package resources

import (
	"encoding/json"
	"time"

	activityModels "dalista/dalista-be/app/repository/models/activityModels"

	"github.com/google/uuid"
)

func idPtrString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// --- Comment ---

type CommentResource struct {
	ID        string  `json:"id"`
	ItemID    string  `json:"item_id"`
	AuthorID  *string `json:"author_id"`
	Body      string  `json:"body"`
	IsSystem  bool    `json:"is_system"`
	CreatedAt string  `json:"created_at"`
}

func CommentFromModel(c *activityModels.Comment) CommentResource {
	return CommentResource{
		ID:        c.ID.String(),
		ItemID:    c.ItemID.String(),
		AuthorID:  idPtrString(c.AuthorID),
		Body:      c.Body,
		IsSystem:  c.IsSystem,
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func CommentsFromModels(cs []activityModels.Comment) []CommentResource {
	out := make([]CommentResource, 0, len(cs))
	for i := range cs {
		out = append(out, CommentFromModel(&cs[i]))
	}
	return out
}

// --- Activity event ---

type ActivityEventResource struct {
	ID         string                 `json:"id"`
	ListID     string                 `json:"list_id"`
	ActorID    *string                `json:"actor_id"`
	Action     string                 `json:"action"`
	TargetType string                 `json:"target_type"`
	TargetID   *string                `json:"target_id"`
	Payload    map[string]interface{} `json:"payload"`
	CreatedAt  string                 `json:"created_at"`
}

func ActivityEventFromModel(e *activityModels.ActivityEvent) ActivityEventResource {
	payload := map[string]interface{}{}
	if e.Payload != "" {
		_ = json.Unmarshal([]byte(e.Payload), &payload)
	}
	return ActivityEventResource{
		ID:         e.ID.String(),
		ListID:     e.ListID.String(),
		ActorID:    idPtrString(e.ActorID),
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   idPtrString(e.TargetID),
		Payload:    payload,
		CreatedAt:  e.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func ActivityEventsFromModels(es []activityModels.ActivityEvent) []ActivityEventResource {
	out := make([]ActivityEventResource, 0, len(es))
	for i := range es {
		out = append(out, ActivityEventFromModel(&es[i]))
	}
	return out
}
