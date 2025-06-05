package scoreboard

import (
	"strings"
	"testing"

	"scoreboard-api/internal"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "Valid name with alphanumeric",
			input:      "Scoreboard123",
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "Valid name with special characters",
			input:      "My_Score-Board 123",
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "Empty name",
			input:      "",
			wantValid:  false,
			wantErrMsg: "Name cannot be empty",
		},
		{
			name:       "Invalid character !",
			input:      "Scoreboard!",
			wantValid:  false,
			wantErrMsg: "Name can only contain alphanumeric characters, hyphens, underscores, and spaces",
		},
		{
			name:       "Invalid character special",
			input:      "Score/board",
			wantValid:  false,
			wantErrMsg: "Name can only contain alphanumeric characters, hyphens, underscores, and spaces",
		},
		{
			name:       "Too long name",
			input:      strings.Repeat("a", 256),
			wantValid:  false,
			wantErrMsg: "Name cannot exceed 255 characters",
		},
		{
			name:       "Max length valid name",
			input:      strings.Repeat("a", 255),
			wantValid:  true,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := internal.NewValidator()
			internal.RegisterCustomValidations(validate)
			nameData := CreateScoreboardPayload{Name: tt.input}
			err := internal.ValidateStruct(validate, nameData)
			if tt.wantValid {
				if err != nil {
					t.Errorf("validateName() unexpected error: %v, want no error", err)
				}
			} else {
				if err == nil {
					t.Errorf("validateName() expected error, got nil")
				} else {
					t.Errorf("validateName() expected error, got %v, want %v", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}
