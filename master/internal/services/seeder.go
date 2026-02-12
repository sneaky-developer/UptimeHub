package services

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// SeedDefaultAdmin creates a default admin user if none exist
func SeedDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.AdminUser{}).Count(&count)
	if count > 0 {
		return
	}

	password := "admin123" // Default password, should be changed
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing default admin password: %v", err)
		return
	}

	admin := models.AdminUser{
		Email:        "admin@uptimehub.local",
		PasswordHash: string(hash),
		Name:         "Admin",
		Role:         "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Error creating default admin: %v", err)
		return
	}

	log.Println("🔑 Default admin created: admin@uptimehub.local / admin123")
}
