package auth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

const adminFile = "data/admin.json"

type Admin struct {
	Username       string `json:"username"`
	PasswordHash   string `json:"password_hash"`
	PasswordChanged bool  `json:"password_changed"`
}

// AdminExists checks if admin account exists
func AdminExists() bool {
	_, err := os.Stat(adminFile)
	return !os.IsNotExist(err)
}

// CreateAdmin creates the default admin account
func CreateAdmin(username, password string) error {
	os.MkdirAll("data", 0755)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := Admin{
		Username:        username,
		PasswordHash:    string(hash),
		PasswordChanged: false, // Force password change on first login
	}

	data, err := json.MarshalIndent(admin, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(adminFile, data, 0600)
}

// ValidateAdmin checks credentials
func ValidateAdmin(username, password string) (*Admin, error) {
	data, err := os.ReadFile(adminFile)
	if err != nil {
		return nil, err
	}

	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return &admin, nil
}

// UpdatePassword changes admin password
func UpdatePassword(newPassword string) error {
	data, err := os.ReadFile(adminFile)
	if err != nil {
		return err
	}

	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin.PasswordHash = string(hash)
	admin.PasswordChanged = true

	newData, err := json.MarshalIndent(admin, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(adminFile, newData, 0600)
}

// GetAdminStatus returns admin account status
func GetAdminStatus() (*Admin, error) {
	data, err := os.ReadFile(adminFile)
	if err != nil {
		return nil, err
	}

	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		return nil, err
	}

	return &admin, nil
}
