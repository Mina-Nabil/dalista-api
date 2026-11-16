package notificationServices

import (
	"context"
	"strings"

	"dalista/dalista-be/app/repository"
	notificationModels "dalista/dalista-be/app/repository/models/notificationModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// validPlatforms is the accepted device_platform set.
var validPlatforms = map[notificationModels.DevicePlatform]bool{
	notificationModels.PlatformIOS:     true,
	notificationModels.PlatformAndroid: true,
	notificationModels.PlatformWeb:     true,
}

// RegisterDevice upserts a push device token for the user. The token is unique
// across users, so a device re-registering (e.g. after switching accounts) is
// reassigned to the current user and its platform/updated_at refreshed.
func RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) (*notificationModels.DeviceToken, error) {
	token = strings.TrimSpace(token)
	p := notificationModels.DevicePlatform(strings.TrimSpace(platform))
	if token == "" || !validPlatforms[p] {
		return nil, ErrValidation
	}

	dt := &notificationModels.DeviceToken{UserID: userID, Token: token, Platform: p}
	err := repository.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "token"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"user_id":    userID,
			"platform":   p,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(dt).Error
	if err != nil {
		return nil, err
	}
	return dt, nil
}

// UnregisterDevice removes a device token for the user (idempotent). Scoped to
// the caller so one user can't delete another's token.
func UnregisterDevice(ctx context.Context, userID uuid.UUID, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrValidation
	}
	return repository.GetDB().WithContext(ctx).
		Where("user_id = ? AND token = ?", userID, token).
		Delete(&notificationModels.DeviceToken{}).Error
}

// tokensForUsers returns all device tokens for the given users, in one query.
func tokensForUsers(ctx context.Context, userIDs []uuid.UUID) ([]notificationModels.DeviceToken, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var tokens []notificationModels.DeviceToken
	err := repository.GetDB().WithContext(ctx).
		Where("user_id IN ?", userIDs).Find(&tokens).Error
	return tokens, err
}
