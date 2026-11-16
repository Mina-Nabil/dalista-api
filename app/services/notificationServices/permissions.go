// Package notificationServices implements notifications: per-user push device
// token registration, sending pushes via FCM (best-effort, with an in-app
// notification log/inbox), the push triggers wired into other features, and the
// admin-initiated SMS send. It is a leaf service — other features call into it
// (sharing, activity); it never calls back into them, to avoid import cycles.
package notificationServices

// Permission strings for notifications.
const (
	// PermManageDevices lets a user register/unregister their own push device tokens.
	PermManageDevices = "notifications:manage_devices"
	// PermReadNotifications lets a user read their own notification inbox.
	PermReadNotifications = "notifications:read"
	// PermSendAdminSMS is admin-only: send an SMS to selected users.
	PermSendAdminSMS = "admin:send_sms"
)

// AllPermissions is every permission string this feature defines.
var AllPermissions = []string{
	PermManageDevices, PermReadNotifications, PermSendAdminSMS,
}

// UserPermissions are the notification permissions a standard user gets by
// default (device management + reading their own inbox; not the admin SMS send).
var UserPermissions = []string{
	PermManageDevices, PermReadNotifications,
}
