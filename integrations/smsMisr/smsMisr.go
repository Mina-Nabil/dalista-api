// Package smsMisr is the SMS Misr client used for OTP delivery (phone sign-in,
// post-signup verification) and, later, admin-initiated SMS. It is intentionally
// small: one Send call. When credentials are not configured (local dev), it logs
// the message instead of calling the provider so phone-auth flows are testable
// offline.
package smsMisr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/config"
)

const (
	// Production SMS endpoint. OTP has a separate template endpoint; a plain
	// message send is sufficient for our OTP + admin-SMS use cases.
	apiURL         = "https://smsmisr.com/api/SMS/"
	requestTimeout = 10 * time.Second
)

// ErrNotConfigured is returned (in non-dev environments) when SMS Misr
// credentials are missing.
type ErrNotConfigured struct{}

func (ErrNotConfigured) Error() string { return "SMS Misr credentials not configured" }

// Send delivers a plaintext SMS to the given E.164 phone number.
//
// Dev convenience: with no credentials configured the message is logged and
// Send returns nil, so OTP flows work locally without a real provider account.
func Send(ctx context.Context, phone, message string) error {
	cfg := config.GetConfig()

	if strings.TrimSpace(cfg.SMSMisrUsername) == "" || strings.TrimSpace(cfg.SMSMisrPassword) == "" {
		if cfg.IsStagingEnvironment() {
			logger.Warning(ctx, "SMS Misr not configured — logging message instead of sending",
				logger.F("phone", phone), logger.F("message", message))
			return nil
		}
		return ErrNotConfigured{}
	}

	// environment: 1 = live, 2 = test. Staging/dev use the test environment.
	environment := "1"
	if cfg.IsStagingEnvironment() {
		environment = "2"
	}

	form := url.Values{
		"environment": {environment},
		"username":    {cfg.SMSMisrUsername},
		"password":    {cfg.SMSMisrPassword},
		"sender":      {cfg.SMSMisrSender},
		"mobile":      {normalizeMobile(phone)},
		"language":    {"1"}, // 1 = English
		"message":     {message},
	}

	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("sms misr: decode response: %w", err)
	}

	// SMS Misr returns code "1901" on a successful send; anything else is a failure.
	if body.Code != "1901" {
		logger.Error(ctx, fmt.Errorf("sms misr send failed: code=%s", body.Code), "SMS Misr send failed",
			logger.F("phone", phone), logger.F("code", body.Code))
		return fmt.Errorf("sms misr: send failed with code %s", body.Code)
	}
	return nil
}

// normalizeMobile strips a leading '+' — SMS Misr expects a bare international
// number (e.g. 2010xxxxxxxx), not E.164 with the plus.
func normalizeMobile(phone string) string {
	return strings.TrimPrefix(strings.TrimSpace(phone), "+")
}
