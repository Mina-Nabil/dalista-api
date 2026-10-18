package forms

// UpdateProfileForm is the body for PUT /profile. All fields are optional; only
// those present are updated. NotificationPrefs is a free-form JSON object.
type UpdateProfileForm struct {
	DisplayName       *string                `json:"display_name" example:"Jane D."`
	PictureURL        *string                `json:"picture_url" example:"https://cdn.example.com/u/jane.png"`
	NotificationPrefs map[string]interface{} `json:"notification_prefs"`
}
