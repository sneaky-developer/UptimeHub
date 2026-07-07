package services

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/config"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// SeedDefaultAdmin creates the initial admin user if none exist.
// Email and password come from ADMIN_EMAIL / ADMIN_PASSWORD. When no password
// is configured, development uses "admin123" for a friction-free first run;
// production generates a random one and prints it once to the logs.
func SeedDefaultAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.AdminUser{}).Count(&count)
	if count > 0 {
		return
	}

	password := cfg.AdminPassword
	generated := false
	if password == "" {
		if cfg.IsDevelopment() {
			password = "admin123"
		} else {
			bytes := make([]byte, 16)
			if _, err := rand.Read(bytes); err != nil {
				log.Printf("Error generating admin password: %v", err)
				return
			}
			password = hex.EncodeToString(bytes)
			generated = true
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing default admin password: %v", err)
		return
	}

	admin := models.AdminUser{
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		Name:         "Admin",
		Role:         "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Error creating default admin: %v", err)
		return
	}

	if generated {
		log.Printf("🔑 Admin user created: %s", cfg.AdminEmail)
		log.Printf("🔑 Generated admin password (shown once, store it now): %s", password)
	} else {
		log.Printf("🔑 Admin user created: %s", cfg.AdminEmail)
	}
}
