// Package cloudinary provides Cloudinary Fetch URL generation with AI smart-cropping
// for vertical (9:16) video feeds.
package cloudinary

import (
	"fmt"
	"net/http"
	"time"
)

var warmUpClient = &http.Client{
	Timeout: 5 * time.Second,
}

// GenerateSmartCropURL builds a Cloudinary Fetch URL that applies g_auto 9:16 crop
// to a YouTube video. Returns "" if cloudName or youtubeID is empty.
func GenerateSmartCropURL(cloudName, youtubeID string) string {
	if cloudName == "" || youtubeID == "" {
		return ""
	}

	youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)

	return fmt.Sprintf(
		"https://res.cloudinary.com/%s/video/fetch/ar_9:16,c_fill,g_auto/%s",
		cloudName,
		youtubeURL,
	)
}

// WarmUp fires a non-blocking HEAD request to trigger Cloudinary's AI processing
// so the video is ready before the user swipes to it.
func WarmUp(url string) {
	if url == "" {
		return
	}

	go func() {
		resp, err := warmUpClient.Head(url)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()
}
