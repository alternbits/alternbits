package root

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dariubs/altern/app/utils"
	"github.com/gin-gonic/gin"
)

const maxScreenshotBytes = 10 * 1024 * 1024

var allowedScreenshotExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func UploadScreenshot(r2 *utils.R2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r2 == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "R2 not configured"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil || file == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedScreenshotExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type — use JPG, PNG, GIF, or WebP"})
			return
		}
		if file.Size > maxScreenshotBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10 MB)"})
			return
		}
		url, err := r2.UploadFile(file, "ai-screenshots")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}
