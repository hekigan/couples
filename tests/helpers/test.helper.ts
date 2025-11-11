import { Page } from '@playwright/test';

/**
 * General test helper functions for E2E tests
 */

/**
 * Create an anonymous user and complete the username setup
 * This is the quickest way to get a logged-in user for testing
 *
 * Flow:
 * 1. Go to home page
 * 2. Click "Play as Guest" button
 * 3. Server creates anonymous user and redirects back to /
 * 4. If redirected to /setup-username, fill in username
 *
 * @param page - Playwright page object
 * @param username - Username to set (defaults to random username)
 * @returns The username that was created
 */
export async function createAnonymousUser(
  page: Page,
  username?: string
): Promise<string> {
  const testUsername = username || `TestUser_${Date.now()}`;

  // Navigate to home page
  await page.goto('/');

  // Find and click the "Play as Guest" button using test ID
  const guestButton = page.getByTestId('play-as-guest-button');

  await guestButton.click();

  // Wait for navigation (server redirects to / or /setup-username)
  await page.waitForLoadState('networkidle');

  // Check if we're on username setup page
  const currentUrl = page.url();
  if (currentUrl.includes('/setup-username')) {
    // Fill in username form
    await page.fill('input[name="username"]', testUsername);
    await page.click('button[type="submit"]');

    // Wait for redirect back to home
    await page.waitForURL('/');
  }

  // Give the page a moment to fully load
  await page.waitForLoadState('networkidle');

  return testUsername;
}

/**
 * Wait for HTMX request to complete and element to update
 * Useful for forms submitted via HTMX
 *
 * @param page - Playwright page object
 * @param selector - Element to wait for
 * @param timeout - Maximum wait time (default: 5000ms)
 */
export async function waitForHTMXRequest(
  page: Page,
  selector: string,
  timeout = 5000
): Promise<void> {
  await Promise.all([
    page.waitForResponse(res => res.status() === 200, { timeout }),
    page.locator(selector).waitFor({ state: 'visible', timeout })
  ]);
}

/**
 * Extract room ID from current URL
 * Useful when you've created a room and need to get its ID
 *
 * @param page - Playwright page object
 * @returns Room ID (UUID string)
 */
export function getRoomIdFromURL(page: Page): string {
  const url = page.url();
  const match = url.match(/\/room\/([a-f0-9-]+)/i);
  if (!match) {
    throw new Error(`Could not extract room ID from URL: ${url}`);
  }
  return match[1];
}

/**
 * Wait for element to contain specific text
 * More reliable than waiting for visibility alone
 *
 * @param page - Playwright page object
 * @param selector - Element selector
 * @param text - Text to wait for
 * @param timeout - Maximum wait time (default: 5000ms)
 */
export async function waitForText(
  page: Page,
  selector: string,
  text: string,
  timeout = 5000
): Promise<void> {
  await page.locator(selector).filter({ hasText: text }).waitFor({
    state: 'visible',
    timeout
  });
}

/**
 * Take a screenshot with a custom name
 * Useful for debugging failed tests
 *
 * @param page - Playwright page object
 * @param name - Screenshot name (without extension)
 */
export async function takeScreenshot(
  page: Page,
  name: string
): Promise<void> {
  await page.screenshot({
    path: `test-results/${name}-${Date.now()}.png`,
    fullPage: true
  });
}

/**
 * Generate a unique room name for testing
 * @returns Unique room name
 */
export function generateRoomName(): string {
  return `Test Room ${Date.now()}`;
}

/**
 * Wait for network idle (useful after form submissions)
 * @param page - Playwright page object
 * @param timeout - Maximum wait time (default: 5000ms)
 */
export async function waitForNetworkIdle(
  page: Page,
  timeout = 5000
): Promise<void> {
  await page.waitForLoadState('networkidle', { timeout });
}
