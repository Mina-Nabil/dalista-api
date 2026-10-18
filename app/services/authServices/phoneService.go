package authServices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dalista/dalista-be/app/repository"
	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/app/utils"
	"dalista/dalista-be/integrations/smsMisr"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	otpTTL         = 10 * time.Minute
	otpMaxAttempts = 5
)

var (
	// ErrOTPNotFound means no active (unconsumed, unexpired) challenge exists.
	ErrOTPNotFound = errors.New("no active verification code")
	// ErrOTPInvalid means the submitted code did not match.
	ErrOTPInvalid = errors.New("invalid verification code")
	// ErrOTPTooManyAttempts means the challenge is locked after too many tries.
	ErrOTPTooManyAttempts = errors.New("too many attempts")
	// ErrPhoneTaken means the phone is already attached to another account.
	ErrPhoneTaken = errors.New("phone already in use")
)

// RequestSigninOTP issues a phone sign-in OTP (no account required).
func RequestSigninOTP(ctx context.Context, phone string) error {
	return issueOTP(ctx, normalizePhone(phone), userModels.PurposeSignin, nil)
}

// RequestVerifyOTP issues a post-signup phone-verification OTP for an existing
// user. Fails if the phone already belongs to a different account.
func RequestVerifyOTP(ctx context.Context, userID uuid.UUID, phone string) error {
	phone = normalizePhone(phone)
	owner, err := GetUserByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if owner != nil && owner.ID != userID {
		return ErrPhoneTaken
	}
	return issueOTP(ctx, phone, userModels.PurposeVerify, &userID)
}

// VerifySigninOTP validates a sign-in OTP and returns the signed-in user (created
// on first use) plus a token pair.
func VerifySigninOTP(ctx context.Context, phone, code string) (*userModels.User, TokenPair, error) {
	phone = normalizePhone(phone)
	if _, err := consumeChallenge(ctx, phone, userModels.PurposeSignin, code); err != nil {
		return nil, TokenPair{}, err
	}

	user, err := GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, TokenPair{}, err
	}
	if user == nil {
		user, err = createUserWithIdentity(ctx, nil, &phone, true,
			userModels.ProviderPhone, phone, nil, nil, nil)
		if err != nil {
			return nil, TokenPair{}, err
		}
	} else {
		if !user.IsActive {
			return nil, TokenPair{}, ErrAccountDeactivated
		}
		if err := ensurePhoneIdentityAndVerified(ctx, user, phone); err != nil {
			return nil, TokenPair{}, err
		}
	}

	CheckAndAttachPendingInvites(ctx, user.ID, phone)

	tokens, err := IssueTokens(user)
	return user, tokens, err
}

// VerifyPhone validates a post-signup verification OTP, marking the user's
// phone verified. Returns the updated user.
func VerifyPhone(ctx context.Context, userID uuid.UUID, phone, code string) (*userModels.User, error) {
	phone = normalizePhone(phone)

	owner, err := GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if owner != nil && owner.ID != userID {
		return nil, ErrPhoneTaken
	}
	if _, err := consumeChallenge(ctx, phone, userModels.PurposeVerify, code); err != nil {
		return nil, err
	}

	user, err := GetUserByID(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if err := ensurePhoneIdentityAndVerified(ctx, user, phone); err != nil {
		return nil, err
	}

	CheckAndAttachPendingInvites(ctx, user.ID, phone)
	return user, nil
}

// issueOTP generates a code, stores its hash, invalidates any prior active
// challenge for the same phone+purpose, and sends the SMS.
func issueOTP(ctx context.Context, phone string, purpose userModels.VerificationPurpose, userID *uuid.UUID) error {
	code, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	db := repository.GetDB().WithContext(ctx)
	now := time.Now()
	// Invalidate prior unconsumed challenges so only the newest code is valid.
	if err := db.Model(&userModels.PhoneVerification{}).
		Where("phone = ? AND purpose = ? AND consumed_at IS NULL", phone, purpose).
		Update("consumed_at", &now).Error; err != nil {
		return err
	}

	challenge := &userModels.PhoneVerification{
		Phone:     phone,
		CodeHash:  utils.HashCode(code),
		Purpose:   purpose,
		UserID:    userID,
		ExpiresAt: now.Add(otpTTL),
	}
	if err := db.Create(challenge).Error; err != nil {
		return err
	}

	msg := fmt.Sprintf("Your Dalista verification code is %s. It expires in %d minutes.", code, int(otpTTL.Minutes()))
	return smsMisr.Send(ctx, phone, msg)
}

// consumeChallenge validates the submitted code against the newest active
// challenge, incrementing attempts on mismatch and consuming it on success.
func consumeChallenge(ctx context.Context, phone string, purpose userModels.VerificationPurpose, code string) (*userModels.PhoneVerification, error) {
	db := repository.GetDB().WithContext(ctx)

	var v userModels.PhoneVerification
	err := db.Where("phone = ? AND purpose = ? AND consumed_at IS NULL", phone, purpose).
		Order("created_at DESC").First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOTPNotFound
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(v.ExpiresAt) {
		return nil, ErrOTPNotFound
	}
	if v.Attempts >= otpMaxAttempts {
		return nil, ErrOTPTooManyAttempts
	}
	if v.CodeHash != utils.HashCode(code) {
		db.Model(&v).Update("attempts", v.Attempts+1)
		return nil, ErrOTPInvalid
	}
	now := time.Now()
	if err := db.Model(&v).Update("consumed_at", &now).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// ensurePhoneIdentityAndVerified sets the user's phone + phone_verified and
// ensures a phone AuthIdentity exists, all in one transaction.
func ensurePhoneIdentityAndVerified(ctx context.Context, user *userModels.User, phone string) error {
	err := repository.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&userModels.User{}).Where("id = ?", user.ID).
			Updates(map[string]interface{}{"phone": phone, "phone_verified": true}).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&userModels.AuthIdentity{}).
			Where("provider = ? AND provider_subject = ?", userModels.ProviderPhone, phone).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Create(&userModels.AuthIdentity{
				UserID:          user.ID,
				Provider:        userModels.ProviderPhone,
				ProviderSubject: phone,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Reflect the committed changes in the caller's in-memory user so the
	// response (and any downstream use) sees the verified phone immediately.
	user.Phone = &phone
	user.PhoneVerified = true
	return nil
}

// normalizePhone trims surrounding whitespace. Full E.164 normalization can be
// layered in later; inputs are validated as e164 at the form boundary.
func normalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}
