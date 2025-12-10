import { Page } from '@playwright/test';

/**
 * Helper functions for testing Server-Sent Events (SSE) in Playwright
 *
 * The Couple Card Game uses SSE extensively for real-time updates via HTMX.
 * These helpers make it easier to wait for and verify SSE events.
 */
export class SSEHelper {
  constructor(private page: Page) {}

  /**
   * Wait for SSE connection to establish
   * @param roomId - The room ID to connect to
   * @param timeout - Maximum time to wait (default: 5000ms)
   */
  async waitForSSEConnection(roomId: string, timeout = 5000): Promise<void> {
    const sseUrl = `/api/v1/stream/rooms/${roomId}/events`;

    await this.page.waitForResponse(
      response => {
        return response.url().includes(sseUrl) && response.status() === 200;
      },
      { timeout }
    );
  }

  /**
   * Wait for a specific SSE event to be received
   * @param eventType - The SSE event type to wait for (e.g., 'player_joined', 'question_drawn')
   * @param timeout - Maximum time to wait (default: 5000ms)
   * @returns Promise that resolves to true if event received, false if timeout
   */
  async waitForSSEEvent(eventType: string, timeout = 5000): Promise<boolean> {
    return new Promise((resolve) => {
      const handler = async (response) => {
        if (response.url().includes('/events')) {
          try {
            const text = await response.text();
            if (text.includes(`event: ${eventType}`)) {
              this.page.off('response', handler);
              resolve(true);
            }
          } catch (error) {
            // Ignore parsing errors
          }
        }
      };

      this.page.on('response', handler);

      // Timeout handler
      setTimeout(() => {
        this.page.off('response', handler);
        resolve(false);
      }, timeout);
    });
  }

  /**
   * Wait for HTMX to swap content after SSE event
   * Useful for verifying that SSE events trigger HTMX updates
   *
   * @param selector - The CSS selector of the element that should be updated
   * @param expectedText - Text that should appear after update (optional)
   * @param timeout - Maximum time to wait (default: 5000ms)
   */
  async waitForHTMXSwap(
    selector: string,
    expectedText?: string,
    timeout = 5000
  ): Promise<void> {
    if (expectedText) {
      await this.page.locator(selector).filter({ hasText: expectedText }).waitFor({
        state: 'visible',
        timeout
      });
    } else {
      await this.page.locator(selector).waitFor({
        state: 'visible',
        timeout
      });
    }
  }

  /**
   * Enable SSE event logging for debugging
   * Call this in your test to see all SSE events in the console
   */
  enableSSELogging(): void {
    this.page.on('response', async (response) => {
      if (response.url().includes('/events')) {
        try {
          const text = await response.text();
          console.log('[SSE Event]:', text);
        } catch (error) {
          // Ignore
        }
      }
    });
  }
}
