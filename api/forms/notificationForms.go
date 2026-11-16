package forms

// RegisterDeviceForm registers a push device token for the caller.
type RegisterDeviceForm struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=ios android web" example:"android"`
}

// UnregisterDeviceForm removes a push device token.
type UnregisterDeviceForm struct {
	Token string `json:"token" binding:"required"`
}

// AdminSMSForm sends an SMS to the selected users (admin).
type AdminSMSForm struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1,dive,uuid"`
	Message string   `json:"message" binding:"required,max=640"`
}
