package authServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	userModels "dalista/dalista-be/app/repository/models/userModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetProfile returns the user and their profile row (creating a blank profile
// if one is somehow missing).
func GetProfile(ctx context.Context, userID uuid.UUID) (*userModels.User, *userModels.UserProfile, error) {
	user, err := GetUserByID(ctx, userID.String())
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, nil
	}
	profile, err := getOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return user, profile, nil
}

// ProfileUpdate carries the mutable profile fields; nil fields are left unchanged.
type ProfileUpdate struct {
	DisplayName       *string
	PictureURL        *string
	NotificationPrefs *string
}

// UpdateProfile applies a partial update and returns the fresh user + profile.
func UpdateProfile(ctx context.Context, userID uuid.UUID, in ProfileUpdate) (*userModels.User, *userModels.UserProfile, error) {
	if _, err := getOrCreateProfile(ctx, userID); err != nil {
		return nil, nil, err
	}

	updates := map[string]interface{}{}
	if in.DisplayName != nil {
		updates["display_name"] = *in.DisplayName
	}
	if in.PictureURL != nil {
		updates["picture_url"] = *in.PictureURL
	}
	if in.NotificationPrefs != nil {
		updates["notification_prefs"] = *in.NotificationPrefs
	}
	if len(updates) > 0 {
		if err := repository.GetDB().WithContext(ctx).
			Model(&userModels.UserProfile{}).Where("user_id = ?", userID).
			Updates(updates).Error; err != nil {
			return nil, nil, err
		}
	}
	return GetProfile(ctx, userID)
}

func getOrCreateProfile(ctx context.Context, userID uuid.UUID) (*userModels.UserProfile, error) {
	db := repository.GetDB().WithContext(ctx)
	var profile userModels.UserProfile
	err := db.Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = userModels.UserProfile{UserID: userID, NotificationPrefs: "{}"}
		if cerr := db.Create(&profile).Error; cerr != nil {
			return nil, cerr
		}
		return &profile, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
