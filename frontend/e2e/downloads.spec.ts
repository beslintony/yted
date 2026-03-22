import { test, expect } from '@playwright/test';
import { createWailsMock, defaultMocks } from './utils/wails-mock';

test.describe('Downloads Page', () => {
  test.beforeEach(async ({ page }) => {
    await createWailsMock(page, defaultMocks);
    await page.goto('/');
    await page.click('[data-testid="downloads-tab"]');
  });

  test('should display download list', async ({ page }) => {
    await expect(page.locator('[data-testid="download-item"]')).toHaveCount(1);
    await expect(page.getByText('Downloading Test Video')).toBeVisible();
  });

  test('should add new download', async ({ page }) => {
    // Fill in URL input
    await page.fill('[data-testid="url-input"]', 'https://youtube.com/watch?v=newvideo');
    
    // Click add button
    await page.click('[data-testid="add-download-button"]');
    
    // Should show success notification or add to list
    await expect(page.getByText('Download added')).toBeVisible();
  });

  test('should validate URL', async ({ page }) => {
    // Fill invalid URL
    await page.fill('[data-testid="url-input"]', 'not-a-valid-url');
    await page.click('[data-testid="add-download-button"]');
    
    // Should show error
    await expect(page.getByText('Invalid URL')).toBeVisible();
  });

  test('should pause and resume download', async ({ page }) => {
    // Click pause button
    await page.click('[data-testid="pause-button"]');
    
    // Status should change to paused
    await expect(page.getByText('Paused')).toBeVisible();
    
    // Click resume button
    await page.click('[data-testid="resume-button"]');
    
    // Status should change back to downloading
    await expect(page.getByText('Downloading')).toBeVisible();
  });

  test('should cancel download', async ({ page }) => {
    // Click cancel button
    await page.click('[data-testid="cancel-button"]');
    
    // Confirm dialog
    await page.click('[data-testid="confirm-cancel"]');
    
    // Download should be removed or marked as cancelled
    await expect(page.getByText('Download cancelled')).toBeVisible();
  });

  test('should clear completed downloads', async ({ page }) => {
    // Click clear completed button
    await page.click('[data-testid="clear-completed-button"]');
    
    // Completed downloads should be removed
    await expect(page.locator('[data-testid="download-item"][data-status="completed"]')).toHaveCount(0);
  });
});
