package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	strongSecret := strings.Repeat("s", 32)

	tests := []struct {
		name    string
		env     string
		secret  string
		wantErr bool
	}{
		{"development allows default secret", "development", "change-me-in-production", false},
		{"development allows empty secret", "development", "", false},
		{"production rejects empty secret", "production", "", true},
		{"production rejects default secret", "production", "change-me-in-production", true},
		{"production rejects compose placeholder", "production", "dev-jwt-secret-change-in-production", true},
		{"production rejects short secret", "production", "short", true},
		{"production accepts strong secret", "production", strongSecret, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AppEnv: tt.env, JWTSecret: tt.secret}
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
