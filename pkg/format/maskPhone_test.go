package format

import (
	"testing"

	"github.com/tj/assert"
)

func TestMaskPhoneNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		skipBeforeMask int
		skipAfterMask  int
		originalNumber string
		expectedResult string
	}{
		{
			name:           "empty phone number",
			skipBeforeMask: 5,
			skipAfterMask:  2,
			originalNumber: "",
			expectedResult: "",
		},
		{
			name:           "common phone number with default mask range",
			skipBeforeMask: 5,
			skipAfterMask:  2,
			originalNumber: "+380661234567",
			expectedResult: "+38066*****67",
		},
		{
			name:           "common phone number with different mask range",
			skipBeforeMask: 3,
			skipAfterMask:  2,
			originalNumber: "+380661234567",
			expectedResult: "+380*******67",
		},
		{
			name:           "phone number with spaces, brackets and dashes",
			skipBeforeMask: 5,
			skipAfterMask:  2,
			originalNumber: "+38 (066) 123-45-67",
			expectedResult: "+38066*****67",
		},
		{
			name:           "phone number with brackets and dashes and without spaces and plus prefix",
			skipBeforeMask: 7,
			skipAfterMask:  3,
			originalNumber: "38(066)-123-45-67",
			expectedResult: "3806612**567",
		},
		{
			name:           "phone number with dashes only",
			skipBeforeMask: 6,
			skipAfterMask:  2,
			originalNumber: "+38-066-123-45-67",
			expectedResult: "+380661****67",
		},
		{
			name:           "phone number with spaces only",
			skipBeforeMask: 5,
			skipAfterMask:  2,
			originalNumber: "+38 066 123 45 67",
			expectedResult: "+38066*****67",
		},
		{
			name:           "phone number without country code",
			skipBeforeMask: 4,
			skipAfterMask:  2,
			originalNumber: "066 123-45-67",
			expectedResult: "0661****67",
		},
		{
			name:           "phone number with cross mask range",
			skipBeforeMask: 10,
			skipAfterMask:  10,
			originalNumber: "+380661234567",
			expectedResult: "+380661234567",
		},
		{
			name:           "phone number with spaces, brackets and dashes and cross mask range",
			skipBeforeMask: 10,
			skipAfterMask:  10,
			originalNumber: "+38 (066) 123-45-67",
			expectedResult: "+380661234567",
		},
		{
			name:           "phone number with spaces, brackets, dashes, cross mask range and without country code",
			skipBeforeMask: 10,
			skipAfterMask:  10,
			originalNumber: "(066) 123-45-67",
			expectedResult: "0661234567",
		},
		{
			name:           "common phone number with skipBefore range only",
			skipBeforeMask: 7,
			skipAfterMask:  0,
			originalNumber: "+380661234567",
			expectedResult: "+3806612*****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := MaskPhoneNumber(tt.originalNumber, PhoneMaskConfig{
				SkipBeforeMask: tt.skipBeforeMask,
				SkipAfterMask:  tt.skipAfterMask,
			})
			assert.Equal(t, tt.expectedResult, res)
		})
	}
}
