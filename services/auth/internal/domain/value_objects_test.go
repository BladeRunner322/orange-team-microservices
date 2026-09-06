package domain

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with dots", "test.user@domain.co.uk", false},
		{"empty", "", true},
		{"missing @", "testexample.com", true},
		{"no domain", "test@", true},
		{"uppercase", "Test@Example.com", false}, // должен привести к нижнему регистру
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// проверяем, что email приведён к нижнему регистру
			if tt.input == "Test@Example.com" && email.String() != "test@example.com" {
				t.Errorf("expected normalized email 'test@example.com', got '%s'", email.String())
			}
		})
	}
}

func TestNewFullName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "John Doe", false},
		{"too short", "A", true},
		{"too long", "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, augue velit cursus nunc,", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFullName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFullName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewPasswordHash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"too short", "abc", true},
		{"valid password len", "1234567890", false},
		{"valid password hash len", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewPasswordHash(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				// Проверяем, что это именно ErrInvalidPasswordHash
				if err != nil && err != ErrInvalidPasswordHash {
					t.Errorf("expected ErrInvalidPasswordHash, got %v", err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(result) == "" {
				t.Errorf("expected non-empty PasswordHash, got empty")
			}
		})
	}
}

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "password123", false},
		{"short", "1234567", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
