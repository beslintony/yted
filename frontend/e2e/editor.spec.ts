import { test, expect } from '@playwright/test';
import { createWailsMock, defaultMocks } from './utils/wails-mock';

test.describe('Editor Page', () => {
  test.beforeEach(async ({ page }) => {
    await createWailsMock(page, defaultMocks);
    await page.goto('/');
    await page.click('[data-testid="editor-tab"]');
  });

  test('should display video selector', async ({ page }) => {
    // Should show video dropdown
    await expect(page.locator('[data-testid="video-select"]')).toBeVisible();
    
    // Open dropdown
    await page.click('[data-testid="video-select"]');
    
    // Should show available videos
    await expect(page.getByText('Test Video 1')).toBeVisible();
    await expect(page.getByText('Test Video 2')).toBeVisible();
  });

  test('should select video and show preview', async ({ page }) => {
    // Select a video
    await page.click('[data-testid="video-select"]');
    await page.click('[data-testid="video-option-test-video-1"]');
    
    // Video player should be visible
    await expect(page.locator('[data-testid="video-player"]')).toBeVisible();
    
    // Video metadata should be displayed
    await expect(page.getByText('1920x1080')).toBeVisible();
    await expect(page.getByText('H264')).toBeVisible();
  });

  test('should show operation tools', async ({ page }) => {
    // Select a video first
    await page.click('[data-testid="video-select"]');
    await page.click('[data-testid="video-option-test-video-1"]');
    
    // Operation tools should be visible
    await expect(page.getByText('Crop & Trim')).toBeVisible();
    await expect(page.getByText('Watermark')).toBeVisible();
    await expect(page.getByText('Convert')).toBeVisible();
    await expect(page.getByText('Effects')).toBeVisible();
  });

  test('should switch between operations', async ({ page }) => {
    // Select a video first
    await page.click('[data-testid="video-select"]');
    await page.click('[data-testid="video-option-test-video-1"]');
    
    // Click on Crop & Trim
    await page.click('[data-testid="operation-crop"]');
    await expect(page.getByText('Crop Settings')).toBeVisible();
    
    // Click on Convert
    await page.click('[data-testid="operation-convert"]');
    await expect(page.getByText('Output Format')).toBeVisible();
  });

  test('should submit edit job', async ({ page }) => {
    // Select a video
    await page.click('[data-testid="video-select"]');
    await page.click('[data-testid="video-option-test-video-1"]');
    
    // Select an operation
    await page.click('[data-testid="operation-convert"]');
    
    // Configure settings
    await page.selectOption('[data-testid="format-select"]', 'mp4');
    
    // Click process button
    await page.click('[data-testid="process-button"]');
    
    // Should show job started
    await expect(page.getByText('Job submitted')).toBeVisible();
  });

  test('should show edit history', async ({ page }) => {
    // Select a video
    await page.click('[data-testid="video-select"]');
    await page.click('[data-testid="video-option-test-video-1"]');
    
    // Click on history tab
    await page.click('[data-testid="history-tab"]');
    
    // History should be visible
    await expect(page.locator('[data-testid="edit-history"]')).toBeVisible();
  });

  test('should show FFmpeg not installed warning', async ({ page }) => {
    // Mock FFmpeg as not installed
    await page.addInitScript(() => {
      // @ts-ignore
      window.go.app.App.CheckFFmpegWithGuidance = async () => ({
        installed: false,
        version: '',
        path: '',
      });
    });
    
    // Reload to apply mock
    await page.reload();
    
    // Should show warning
    await expect(page.getByText('FFmpeg not installed')).toBeVisible();
    await expect(page.getByText('Install Now')).toBeVisible();
  });
});
