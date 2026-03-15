package app

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
	applog "yted/internal/log"
)

// VerifyAndRepairDownloads checks all incomplete downloads and updates their status
// if the files already exist (downloads that completed but weren't marked correctly)
func (a *App) VerifyAndRepairDownloads() error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	// Get all downloads (including completed) to check for inconsistencies
	downloads, err := a.db.ListDownloads()
	if err != nil {
		logger.Error("Download", "Failed to list downloads for verification", err)
		return err
	}

	repaired := 0
	for _, dl := range downloads {
		// Skip if already marked as completed
		if dl.Status == "completed" {
			continue
		}

		// Try to find the file for this download
		if dl.Title != nil && a.config != nil {
			downloadPath := a.config.Get().DownloadPath
			youtubeID := extractYoutubeID(dl.URL)
			
			ext := "mp4"
			if dl.Quality != nil && *dl.Quality == "audio" {
				ext = "mp3"
			}
			
			// Look for the file
			foundPath := findDownloadedFile(downloadPath, youtubeID, ext)
			
			if foundPath != "" {
				// File exists! Mark as completed
				logger.Info("Download", "Found existing file for incomplete download, marking as completed", map[string]string{
					"id":       dl.ID,
					"file":     foundPath,
					"previous": dl.Status,
				})
				
				if err := a.db.CompleteDownload(dl.ID); err != nil {
					logger.Error("Download", "Failed to repair download status", err, map[string]string{"id": dl.ID})
				} else {
					repaired++
					// Emit event to notify frontend
					runtime.EventsEmit(a.ctx, "download:completed", dl.ID)
				}
			}
		}
	}

	if repaired > 0 {
		logger.Info("Download", "Repaired download statuses", map[string]int{
			"count": repaired,
		})
	}

	return nil
}

// CheckDownloadStatus checks if a specific download's file exists
func (a *App) CheckDownloadStatus(downloadID string) (map[string]interface{}, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil, nil
	}

	dl, err := a.db.GetDownload(downloadID)
	if err != nil {
		return nil, err
	}
	if dl == nil {
		return map[string]interface{}{
			"exists":   false,
			"status":   "not_found",
			"filePath": "",
		}, nil
	}

	// Try to find the file
	var filePath string
	if a.config != nil {
		downloadPath := a.config.Get().DownloadPath
		youtubeID := extractYoutubeID(dl.URL)
		
		ext := "mp4"
		if dl.Quality != nil && *dl.Quality == "audio" {
			ext = "mp3"
		}
		
		filePath = findDownloadedFile(downloadPath, youtubeID, ext)
	}

	result := map[string]interface{}{
		"exists":   filePath != "",
		"status":   dl.Status,
		"filePath": filePath,
		"progress": dl.Progress,
	}

	// If file exists but status isn't completed, offer to repair
	if filePath != "" && dl.Status != "completed" {
		logger.Info("Download", "File exists but status is not completed", map[string]string{
			"id":     downloadID,
			"status": dl.Status,
			"file":   filePath,
		})
		result["needsRepair"] = true
	}

	return result, nil
}

// SyncDownloadWithFile checks if the file exists and updates download status accordingly
func (a *App) SyncDownloadWithFile(downloadID string) error {
	logger := applog.GetLogger()

	status, err := a.CheckDownloadStatus(downloadID)
	if err != nil {
		return err
	}

	if status["needsRepair"] == true {
		logger.Info("Download", "Syncing download status with file existence", map[string]string{
			"id": downloadID,
		})
		
		if err := a.db.CompleteDownload(downloadID); err != nil {
			logger.Error("Download", "Failed to sync download status", err, map[string]string{"id": downloadID})
			return err
		}
		
		// Emit completion event
		runtime.EventsEmit(a.ctx, "download:completed", downloadID)
	}

	return nil
}
