package sharingServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SharedList is a list shared with the caller, plus the caller's role on it.
type SharedList struct {
	List listModels.List
	Role sharingModels.Role
}

// ListSharedWithMe returns the lists shared with the caller (not owned) + role.
func ListSharedWithMe(ctx context.Context, userID uuid.UUID) ([]SharedList, error) {
	var shares []sharingModels.ListShare
	if err := repository.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).Order("created_at DESC").Find(&shares).Error; err != nil {
		return nil, err
	}
	out := make([]SharedList, 0, len(shares))
	for _, s := range shares {
		var list listModels.List
		if err := repository.GetDB().WithContext(ctx).Where("id = ?", s.ListID).First(&list).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		out = append(out, SharedList{List: list, Role: s.Role})
	}
	return out, nil
}

// Collaborator is one active collaborator on a list.
type Collaborator struct {
	UserID string `gorm:"column:user_id" json:"user_id"`
	Email  string `gorm:"column:email" json:"email"`
	Phone  string `gorm:"column:phone" json:"phone"`
	Role   string `gorm:"column:role" json:"role"`
}

// Invite is one open pending invite on a list.
type Invite struct {
	ID    string `json:"id"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

// Collaborators is the collaborator listing for a list: active collaborators
// (visible to any member) plus open invites (owner only).
type Collaborators struct {
	Collaborators []Collaborator `json:"collaborators"`
	Invites       []Invite       `json:"invites"`
}

// ListCollaborators returns a list's collaborators. Requires the caller to be a
// member (owner/editor/viewer) → else ErrNotFound. Pending invites are included
// only for the owner.
func ListCollaborators(ctx context.Context, userID, listID uuid.UUID) (*Collaborators, error) {
	db := repository.GetDB().WithContext(ctx)

	var list listModels.List
	if err := db.Where("id = ?", listID).First(&list).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	isOwner := list.OwnerID == userID
	if !isOwner {
		var member sharingModels.ListShare
		if err := db.Where("list_id = ? AND user_id = ?", listID, userID).First(&member).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
	}

	out := &Collaborators{Collaborators: []Collaborator{}, Invites: []Invite{}}
	if err := db.Table("list_shares s").
		Select(`s.user_id AS user_id, COALESCE(u.email,'') AS email, COALESCE(u.phone,'') AS phone, s.role AS role`).
		Joins("JOIN users u ON u.id = s.user_id").
		Where("s.list_id = ?", listID).
		Order("s.created_at ASC").Scan(&out.Collaborators).Error; err != nil {
		return nil, err
	}

	if isOwner {
		var invites []sharingModels.PendingInvite
		if err := db.Where("list_id = ? AND consumed_at IS NULL", listID).
			Order("created_at ASC").Find(&invites).Error; err != nil {
			return nil, err
		}
		for _, inv := range invites {
			out.Invites = append(out.Invites, Invite{ID: inv.ID.String(), Phone: inv.Phone, Role: string(inv.Role)})
		}
	}
	return out, nil
}
