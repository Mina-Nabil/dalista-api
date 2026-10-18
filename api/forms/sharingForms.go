package forms

// ShareListForm shares a list with a phone number at a role.
type ShareListForm struct {
	Phone string `json:"phone" binding:"required,e164" example:"+201001234567"`
	Role  string `json:"role" binding:"required,oneof=editor viewer" example:"editor"`
}

// ChangeRoleForm changes a collaborator's role.
type ChangeRoleForm struct {
	Role string `json:"role" binding:"required,oneof=editor viewer" example:"viewer"`
}
