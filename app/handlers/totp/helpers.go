package totp

import (
	"github.com/dariubs/altern/app/models"
	totplib "github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// NewSecret generates a fresh TOTP secret for the user and saves it to the DB.
// Does not enable TOTP — caller must call EnableTOTP after the user confirms.
func NewSecret(db *gorm.DB, user *models.User) error {
	key, err := totplib.Generate(totplib.GenerateOpts{
		Issuer:      issuer(),
		AccountName: accountLabel(user),
	})
	if err != nil {
		return err
	}
	user.TOTPSecret = key.Secret()
	return db.Save(user).Error
}

// QRBase64 renders the user's current TOTP secret as a base64-encoded PNG QR code.
func QRBase64(user *models.User) (string, error) {
	return qrCodePNGBase64(user)
}

// GenerateCodes generates n recovery codes, returning plain-text and hashed versions.
func GenerateCodes(n int) (plain []string, hashed []string, err error) {
	return generateRecoveryCodes(n)
}

// ValidateCode checks whether a 6-digit TOTP code is valid for the given secret.
func ValidateCode(code, secret string) bool {
	return totplib.Validate(code, secret)
}

// ConsumeCode removes a matching recovery code from the user's list (single-use).
func ConsumeCode(user *models.User, code string) (string, bool) {
	return consumeRecoveryCode(user, code)
}
