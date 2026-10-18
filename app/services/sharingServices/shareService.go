package sharingServices

import (
	"context"
	"errors"
	"strings"

	"dalista/dalista-be/app/repository"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"
	userModels "dalista/dalista-be/app/repository/models/userModels"
	billingServices "dalista/dalista-be/app/services/billingServices"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShareKind distinguishes an immediate share (registered user) from a pending
// invite (phone with no account yet).
type ShareKind string

const (
	KindShare  ShareKind = "share"
	KindInvite ShareKind = "invite"
)

// ShareResult reports the outcome of ShareListByPhone.
type ShareResult struct {
	Kind   ShareKind
	Share  *sharingModels.ListShare
	Invite *sharingModels.PendingInvite
}

func validRole(r string) (sharingModels.Role, bool) {
	switch sharingModels.Role(r) {
	case sharingModels.RoleEditor:
		return sharingModels.RoleEditor, true
	case sharingModels.RoleViewer:
		return sharingModels.RoleViewer, true
	default:
		return "", false
	}
}

// ShareListByPhone shares a list (owner-only) with a phone number at a role. If
// the number belongs to a registered user, a ListShare is created immediately;
// otherwise a PendingInvite is recorded and auto-attached when that number is
// verified. Enforces the owner's tier per-list collaborator cap.
func ShareListByPhone(ctx context.Context, ownerID, listID uuid.UUID, phone, roleStr string) (*ShareResult, error) {
	role, ok := validRole(roleStr)
	if !ok {
		return nil, ErrValidation
	}
	if _, err := loadOwnedList(ctx, ownerID, listID); err != nil {
		return nil, err
	}
	phone = strings.TrimSpace(phone)

	target, err := findUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if target != nil && target.ID == ownerID {
		return nil, ErrSelfShare
	}

	db := repository.GetDB().WithContext(ctx)

	if target != nil {
		// Existing registered user → immediate share (idempotent role update).
		var existing sharingModels.ListShare
		err := db.Where("list_id = ? AND user_id = ?", listID, target.ID).First(&existing).Error
		if err == nil {
			existing.Role = role
			if uerr := db.Model(&existing).Update("role", role).Error; uerr != nil {
				return nil, uerr
			}
			return &ShareResult{Kind: KindShare, Share: &existing}, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := enforceCap(ctx, ownerID, listID); err != nil {
			return nil, err
		}
		share := &sharingModels.ListShare{ListID: listID, UserID: target.ID, Role: role, InvitedBy: ownerID}
		if err := db.Create(share).Error; err != nil {
			return nil, err
		}
		return &ShareResult{Kind: KindShare, Share: share}, nil
	}

	// No account yet → pending invite (idempotent role update on the open invite).
	var invite sharingModels.PendingInvite
	err = db.Where("list_id = ? AND phone = ? AND consumed_at IS NULL", listID, phone).First(&invite).Error
	if err == nil {
		if uerr := db.Model(&invite).Update("role", role).Error; uerr != nil {
			return nil, uerr
		}
		invite.Role = role
		return &ShareResult{Kind: KindInvite, Invite: &invite}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := enforceCap(ctx, ownerID, listID); err != nil {
		return nil, err
	}
	newInvite := &sharingModels.PendingInvite{ListID: listID, Phone: phone, Role: role, InvitedBy: ownerID}
	if err := db.Create(newInvite).Error; err != nil {
		return nil, err
	}
	return &ShareResult{Kind: KindInvite, Invite: newInvite}, nil
}

// ChangeRole updates a collaborator's role (owner-only).
func ChangeRole(ctx context.Context, ownerID, listID, targetUserID uuid.UUID, roleStr string) error {
	role, ok := validRole(roleStr)
	if !ok {
		return ErrValidation
	}
	if _, err := loadOwnedList(ctx, ownerID, listID); err != nil {
		return err
	}
	res := repository.GetDB().WithContext(ctx).Model(&sharingModels.ListShare{}).
		Where("list_id = ? AND user_id = ?", listID, targetUserID).Update("role", role)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Revoke removes a collaborator from a list (owner-only).
func Revoke(ctx context.Context, ownerID, listID, targetUserID uuid.UUID) error {
	if _, err := loadOwnedList(ctx, ownerID, listID); err != nil {
		return err
	}
	res := repository.GetDB().WithContext(ctx).
		Where("list_id = ? AND user_id = ?", listID, targetUserID).Delete(&sharingModels.ListShare{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CancelInvite deletes a pending invite on a list (owner-only).
func CancelInvite(ctx context.Context, ownerID, listID, inviteID uuid.UUID) error {
	if _, err := loadOwnedList(ctx, ownerID, listID); err != nil {
		return err
	}
	res := repository.GetDB().WithContext(ctx).
		Where("id = ? AND list_id = ?", inviteID, listID).Delete(&sharingModels.PendingInvite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// --- helpers ---

// loadOwnedList loads a list and confirms the caller owns it; otherwise
// ErrNotFound (sharing management is owner-only and existence is not leaked).
func loadOwnedList(ctx context.Context, ownerID, listID uuid.UUID) (*listModels.List, error) {
	var list listModels.List
	err := repository.GetDB().WithContext(ctx).Where("id = ?", listID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if list.OwnerID != ownerID {
		return nil, ErrNotFound
	}
	return &list, nil
}

func findUserByPhone(ctx context.Context, phone string) (*userModels.User, error) {
	var user userModels.User
	err := repository.GetDB().WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// enforceCap returns ErrCapReached when the list already has the owner's tier
// cap of collaborators (active shares + open invites). Unlimited (-1) skips.
func enforceCap(ctx context.Context, ownerID, listID uuid.UUID) error {
	_, limits, err := billingServices.GetUserLimits(ctx, ownerID)
	if err != nil {
		return err
	}
	if limits.PerListUserCap == billingServices.Unlimited {
		return nil
	}
	count, err := collaboratorCount(ctx, listID)
	if err != nil {
		return err
	}
	if count >= int64(limits.PerListUserCap) {
		return ErrCapReached
	}
	return nil
}

// collaboratorCount is active shares + open (unconsumed) pending invites.
func collaboratorCount(ctx context.Context, listID uuid.UUID) (int64, error) {
	db := repository.GetDB().WithContext(ctx)
	var shares, invites int64
	if err := db.Model(&sharingModels.ListShare{}).Where("list_id = ?", listID).Count(&shares).Error; err != nil {
		return 0, err
	}
	if err := db.Model(&sharingModels.PendingInvite{}).
		Where("list_id = ? AND consumed_at IS NULL", listID).Count(&invites).Error; err != nil {
		return 0, err
	}
	return shares + invites, nil
}
