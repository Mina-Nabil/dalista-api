package resources

import (
	"encoding/json"

	userModels "dalista/dalista-be/app/repository/models/userModels"
)

// ProfileResource is the merged user + profile view returned by the profile endpoints.
type ProfileResource struct {
	ID                string                 `json:"id"`
	Email             *string                `json:"email"`
	Phone             *string                `json:"phone"`
	PhoneVerified     bool                   `json:"phone_verified"`
	DisplayName       *string                `json:"display_name"`
	PictureURL        *string                `json:"picture_url"`
	NotificationPrefs map[string]interface{} `json:"notification_prefs"`
}

// ProfileFromModel merges a user and its profile into one resource.
func ProfileFromModel(u *userModels.User, p *userModels.UserProfile) ProfileResource {
	prefs := map[string]interface{}{}
	if p != nil && p.NotificationPrefs != "" {
		_ = json.Unmarshal([]byte(p.NotificationPrefs), &prefs)
	}
	res := ProfileResource{
		ID:                u.ID.String(),
		Email:             u.Email,
		Phone:             u.Phone,
		PhoneVerified:     u.PhoneVerified,
		NotificationPrefs: prefs,
	}
	if p != nil {
		res.DisplayName = p.DisplayName
		res.PictureURL = p.PictureURL
	}
	return res
}
