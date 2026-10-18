package authServices

import (
	"context"
	"errors"

	"dalista/dalista-be/app/repository"
	userModels "dalista/dalista-be/app/repository/models/userModels"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetUserByID returns the user, or (nil, nil) if not found.
func GetUserByID(ctx context.Context, id string) (*userModels.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, nil
	}
	var user userModels.User
	err = repository.GetDB().WithContext(ctx).Where("id = ?", uid).First(&user).Error
	return firstResult(&user, err)
}

// GetUserByEmail returns the user with the given email, or (nil, nil).
func GetUserByEmail(ctx context.Context, email string) (*userModels.User, error) {
	var user userModels.User
	err := repository.GetDB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	return firstResult(&user, err)
}

// GetUserByPhone returns the user with the given phone, or (nil, nil).
func GetUserByPhone(ctx context.Context, phone string) (*userModels.User, error) {
	var user userModels.User
	err := repository.GetDB().WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return firstResult(&user, err)
}

// GetIdentity returns the auth identity for (provider, subject), or (nil, nil).
func GetIdentity(ctx context.Context, provider userModels.Provider, subject string) (*userModels.AuthIdentity, error) {
	var id userModels.AuthIdentity
	err := repository.GetDB().WithContext(ctx).
		Where("provider = ? AND provider_subject = ?", provider, subject).First(&id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// SetUserActive flips the is_active flag (used by admin suspend/reactivate later).
func SetUserActive(ctx context.Context, userID uuid.UUID, active bool) error {
	return repository.GetDB().WithContext(ctx).
		Model(&userModels.User{}).Where("id = ?", userID).
		Update("is_active", active).Error
}

func firstResult(user *userModels.User, err error) (*userModels.User, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// createUserWithIdentity creates a user, its blank profile, and one auth
// identity atomically. Used by every signup path (email/phone/social).
// `email` and `phone` may be nil depending on the path. displayName/picture are
// seeded into the profile when provided.
func createUserWithIdentity(
	ctx context.Context,
	email, phone *string,
	phoneVerified bool,
	provider userModels.Provider,
	subject string,
	passwordHash *string,
	displayName, pictureURL *string,
) (*userModels.User, error) {
	user := &userModels.User{
		Email:         email,
		Phone:         phone,
		Role:          userModels.RoleUser,
		PhoneVerified: phoneVerified,
		IsActive:      true,
	}

	err := repository.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		identity := &userModels.AuthIdentity{
			UserID:          user.ID,
			Provider:        provider,
			ProviderSubject: subject,
			PasswordHash:    passwordHash,
		}
		if err := tx.Create(identity).Error; err != nil {
			return err
		}
		profile := &userModels.UserProfile{
			UserID:            user.ID,
			DisplayName:       displayName,
			PictureURL:        pictureURL,
			NotificationPrefs: "{}",
		}
		return tx.Create(profile).Error
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// linkIdentity attaches a new provider identity to an existing user.
func linkIdentity(ctx context.Context, userID uuid.UUID, provider userModels.Provider, subject string, passwordHash *string) error {
	identity := &userModels.AuthIdentity{
		UserID:          userID,
		Provider:        provider,
		ProviderSubject: subject,
		PasswordHash:    passwordHash,
	}
	return repository.GetDB().WithContext(ctx).Create(identity).Error
}
