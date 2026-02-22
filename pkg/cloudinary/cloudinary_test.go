package cloudinary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSmartCropURL(t *testing.T) {
	tests := map[string]struct {
		cloudName string
		youtubeID string
		expected  string
	}{
		"valid input": {
			"my_cloud",
			"abc123",
			"https://res.cloudinary.com/my_cloud/video/fetch/ar_9:16,c_fill,g_auto/https://www.youtube.com/watch?v=abc123",
		},
		"empty cloud name": {
			"",
			"abc123",
			"",
		},
		"empty youtube ID": {
			"my_cloud",
			"",
			"",
		},
		"both empty": {
			"",
			"",
			"",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := GenerateSmartCropURL(tt.cloudName, tt.youtubeID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWarmUp_EmptyURL(t *testing.T) {
	// Should not panic on empty URL.
	WarmUp("")
}
