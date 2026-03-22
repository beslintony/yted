package editor

import (
	"context"
	"sync"
	"time"

	"yted/internal/db"
	"yted/internal/log"
)

// EditTask represents a task in the edit queue
type EditTask struct {
	JobID     string
	Video     *db.Video
	Operation string
	Settings  db.EditSettings
	Editor    *Editor
}

// EditQueue manages the edit job queue
type EditQueue struct {
	tasks   chan *EditTask
	workers int
	wg      sync.WaitGroup
	logger  *log.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewQueue creates a new edit queue with the specified number of workers
func NewQueue(workers int) *EditQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &EditQueue{
		tasks:   make(chan *EditTask, 100),
		workers: workers,
		logger:  log.GetLogger(),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the worker goroutines
func (q *EditQueue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	q.logger.Info("Editor", "Edit queue started", map[string]int{"workers": q.workers})
}

// Stop stops the queue and waits for workers to finish
func (q *EditQueue) Stop() {
	q.cancel()
	close(q.tasks)
	q.wg.Wait()
	q.logger.Info("Editor", "Edit queue stopped")
}

// Submit adds a task to the queue
func (q *EditQueue) Submit(task *EditTask) error {
	select {
	case q.tasks <- task:
		return nil
	case <-q.ctx.Done():
		return context.Canceled
	}
}

// worker processes tasks from the queue
func (q *EditQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case task, ok := <-q.tasks:
			if !ok {
				return
			}
			q.processTask(task)
		case <-q.ctx.Done():
			return
		}
	}
}

// processTask processes a single edit task
func (q *EditQueue) processTask(task *EditTask) {
	logger := q.logger
	logger.Info("Editor", "Processing edit task", map[string]string{
		"job_id":    task.JobID,
		"operation": task.Operation,
		"video":     task.Video.Title,
	})

	// Update status to processing
	if err := task.Editor.db.UpdateEditJob(&db.EditJob{
		ID:       task.JobID,
		Status:   "processing",
		Progress: 0,
	}); err != nil {
		logger.Error("Editor", "Failed to update job status", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Register job
	task.Editor.mu.Lock()
	task.Editor.activeJobs[task.JobID] = &EditJob{
		ID:        task.JobID,
		VideoID:   task.Video.ID,
		Operation: task.Operation,
		Settings:  task.Settings,
		Status:    "processing",
		Cancel:    cancel,
	}
	task.Editor.mu.Unlock()

	// Progress callback
	progressFn := func(progress float64, eta string) {
		task.Editor.mu.Lock()
		if job, exists := task.Editor.activeJobs[task.JobID]; exists {
			job.Progress = progress
		}
		task.Editor.mu.Unlock()

		// Update database (throttle updates)
		if int(progress*100)%5 == 0 { // Every 5%
			_ = task.Editor.db.UpdateEditJobProgress(task.JobID, progress)
		}
	}

	// Execute the edit operation
	outputPath, err := q.executeOperation(ctx, task, progressFn)

	// Unregister job
	task.Editor.mu.Lock()
	delete(task.Editor.activeJobs, task.JobID)
	task.Editor.mu.Unlock()

	// Update final status
	if err != nil {
		logger.Error("Editor", "Edit operation failed", err, map[string]string{
			"job_id":    task.JobID,
			"operation": task.Operation,
		})
		if err := task.Editor.db.FailEditJob(task.JobID, err.Error()); err != nil {
			logger.Error("Editor", "Failed to mark job as failed", err)
		}
	} else {
		logger.Info("Editor", "Edit operation completed", map[string]string{
			"job_id":      task.JobID,
			"operation":   task.Operation,
			"output_path": outputPath,
		})

		// Create output video record
		outputVideo := &db.Video{
			ID:           task.JobID + "_output",
			YoutubeID:    task.Video.YoutubeID,
			Title:        task.Video.Title + " (edited)",
			Channel:      task.Video.Channel,
			ChannelID:    task.Video.ChannelID,
			Description:  task.Video.Description,
			ThumbnailURL: task.Video.ThumbnailURL,
			FilePath:     outputPath,
			FileHash:     task.Video.FileHash + "_edited_" + task.JobID,
			IsManaged:    task.Video.IsManaged,
			DownloadedAt: time.Now(),
		}

		if err := task.Editor.db.CreateVideo(outputVideo); err != nil {
			logger.Error("Editor", "Failed to create output video record", err)
		} else {
			if err := task.Editor.db.CompleteEditJob(task.JobID, outputVideo.ID); err != nil {
				logger.Error("Editor", "Failed to mark job as completed", err)
			}
		}
	}
}

// executeOperation executes the specified edit operation
func (q *EditQueue) executeOperation(ctx context.Context, task *EditTask, progressFn ProgressCallback) (string, error) {
	switch task.Operation {
	case "crop":
		return task.Editor.executeCrop(ctx, task.Video, task.Settings, progressFn)
	case "watermark":
		return task.Editor.executeWatermark(ctx, task.Video, task.Settings, progressFn)
	case "convert":
		return task.Editor.executeConvert(ctx, task.Video, task.Settings, progressFn)
	case "effects":
		return task.Editor.executeEffects(ctx, task.Video, task.Settings, progressFn)
	case "combine":
		return task.Editor.executeCombine(ctx, task.Video, task.Settings, progressFn)
	default:
		return "", fmt.Errorf("unknown operation: %s", task.Operation)
	}
}
