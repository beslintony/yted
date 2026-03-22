import { test, expect } from '@playwright/test';
import { createWailsMock, defaultMocks } from './utils/wails-mock';

test.describe('Library Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock Wails runtime before navigating
    await createWailsMock(page, defaultMocks);
    await page.goto('/');
    
    // Navigate to library tab
    await page.click('[data-testid="library-tab"]');
  });

  test('should display video list', async ({ page }) => {
    // Check that videos are displayed
    await expect(page.locator('[data-testid="video-card"]')).toHaveCount(2);
    
    // Check video titles are visible
    await expect(page.getByText('Test Video 1')).toBeVisible();
    await expect(page.getByText('Test Video 2')).toBeVisible();
  });

  test('should search videos', async ({ page }) => {
    // Type in search box
    await page.fill('[data-testid="search-input"]', 'Video 1');
    
    // Should only show matching video
    await expect(page.locator('[data-testid="video-card"]')).toHaveCount(1);
    await expect(page.getByText('Test Video 1')).toBeVisible();
    await expect(page.getByText('Test Video 2')).not.toBeVisible();
  });

  test('should switch between grid and list view', async ({ page }) => {
    // Default is grid view
    await expect(page.locator('[data-testid="grid-view"]')).toBeVisible();
    
    // Switch to list view
    await page.click('[data-testid="list-view-button"]');
    await expect(page.locator('[data-testid="list-view"]')).toBeVisible();
    
    // Switch back to grid view
    await page.click('[data-testid="grid-view-button"]');
    await expect(page.locator('[data-testid="grid-view"]')).toBeVisible();
  });

  test('should navigate to video player', async ({ page }) => {
    // Click on first video
    await page.click('[data-testid="video-card"]:first-child');
    
    // Should navigate to video detail or player
    await expect(page.locator('[data-testid="video-player"]')).toBeVisible();
  });

  test('should show video metadata', async ({ page }) => {
    // Check video metadata is displayed
    await expect(page.getByText('Test Channel')).toBeVisible();
    await expect(page.getByText('1080p')).toBeVisible();
    await expect(page.getByText('5:00')).toBeVisible(); // 300 seconds formatted
  });
});
