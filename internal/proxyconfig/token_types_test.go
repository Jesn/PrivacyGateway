package proxyconfig

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAccessToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{
			name:      "no expiration",
			expiresAt: nil,
			want:      false,
		},
		{
			name:      "future expiration",
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			want:      false,
		},
		{
			name:      "past expiration",
			expiresAt: timePtr(time.Now().Add(-time.Hour)),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AccessToken{
				ExpiresAt: tt.expiresAt,
			}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("AccessToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessToken_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		expiresAt *time.Time
		want      bool
	}{
		{
			name:      "enabled and not expired",
			enabled:   true,
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			want:      true,
		},
		{
			name:      "disabled",
			enabled:   false,
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			want:      false,
		},
		{
			name:      "enabled but expired",
			enabled:   true,
			expiresAt: timePtr(time.Now().Add(-time.Hour)),
			want:      false,
		},
		{
			name:      "enabled and no expiration",
			enabled:   true,
			expiresAt: nil,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AccessToken{
				Enabled:   tt.enabled,
				ExpiresAt: tt.expiresAt,
			}
			if got := token.IsActive(); got != tt.want {
				t.Errorf("AccessToken.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessToken_GetStatus(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		expiresAt *time.Time
		want      string
	}{
		{
			name:      "active token",
			enabled:   true,
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			want:      TokenStatusActive,
		},
		{
			name:      "disabled token",
			enabled:   false,
			expiresAt: timePtr(time.Now().Add(time.Hour)),
			want:      TokenStatusDisabled,
		},
		{
			name:      "expired token",
			enabled:   true,
			expiresAt: timePtr(time.Now().Add(-time.Hour)),
			want:      TokenStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AccessToken{
				Enabled:   tt.enabled,
				ExpiresAt: tt.expiresAt,
			}
			if got := token.GetStatus(); got != tt.want {
				t.Errorf("AccessToken.GetStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessToken_UpdateUsage(t *testing.T) {
	token := &AccessToken{
		UsageCount: 5,
		LastUsed:   nil,
	}

	beforeUpdate := time.Now()
	token.UpdateUsage()
	afterUpdate := time.Now()

	if token.UsageCount != 6 {
		t.Errorf("Expected usage count to be 6, got %d", token.UsageCount)
	}

	if token.LastUsed == nil {
		t.Error("Expected LastUsed to be set")
	} else if token.LastUsed.Before(beforeUpdate) || token.LastUsed.After(afterUpdate) {
		t.Error("LastUsed time is not within expected range")
	}

	if token.UpdatedAt.Before(beforeUpdate) || token.UpdatedAt.After(afterUpdate) {
		t.Error("UpdatedAt time is not within expected range")
	}
}

func TestAccessToken_Validate(t *testing.T) {
	tests := []struct {
		name    string
		token   *AccessToken
		wantErr bool
		errType error
	}{
		{
			name: "valid token",
			token: &AccessToken{
				Name:      "test-token",
				TokenHash: "hash123",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			token: &AccessToken{
				Name:      "",
				TokenHash: "hash123",
			},
			wantErr: true,
			errType: ErrTokenNameRequired,
		},
		{
			name: "name too long",
			token: &AccessToken{
				Name:      string(make([]byte, MaxTokenNameLength+1)),
				TokenHash: "hash123",
			},
			wantErr: true,
			errType: ErrTokenNameTooLong,
		},
		{
			name: "empty token hash",
			token: &AccessToken{
				Name:      "test-token",
				TokenHash: "",
			},
			wantErr: true,
			errType: ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.token.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AccessToken.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("AccessToken.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *TokenCreateRequest
		wantErr bool
		errType error
	}{
		{
			name: "valid request",
			req: &TokenCreateRequest{
				Name:        "test-token",
				Description: "test description",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: &TokenCreateRequest{
				Name: "",
			},
			wantErr: true,
			errType: ErrTokenNameRequired,
		},
		{
			name: "name too long",
			req: &TokenCreateRequest{
				Name: string(make([]byte, MaxTokenNameLength+1)),
			},
			wantErr: true,
			errType: ErrTokenNameTooLong,
		},
		{
			name: "past expiration time",
			req: &TokenCreateRequest{
				Name:      "test-token",
				ExpiresAt: timePtr(time.Now().Add(-time.Hour)),
			},
			wantErr: true,
		},
		{
			name: "future expiration time",
			req: &TokenCreateRequest{
				Name:      "test-token",
				ExpiresAt: timePtr(time.Now().Add(time.Hour)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("ValidateCreateRequest() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestTokenSerialization(t *testing.T) {
	now := time.Now()
	token := &AccessToken{
		ID:          "test-id",
		Name:        "test-token",
		TokenHash:   "hash123",
		ExpiresAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastUsed:    &now,
		UsageCount:  10,
		Enabled:     true,
		CreatedBy:   "admin",
		Description: "test description",
	}

	// 测试序列化
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}

	// 测试反序列化
	var unmarshaled AccessToken
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal token: %v", err)
	}

	// 验证字段
	if unmarshaled.ID != token.ID {
		t.Errorf("ID mismatch: got %v, want %v", unmarshaled.ID, token.ID)
	}
	if unmarshaled.Name != token.Name {
		t.Errorf("Name mismatch: got %v, want %v", unmarshaled.Name, token.Name)
	}
	if unmarshaled.UsageCount != token.UsageCount {
		t.Errorf("UsageCount mismatch: got %v, want %v", unmarshaled.UsageCount, token.UsageCount)
	}
}

// 辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}
