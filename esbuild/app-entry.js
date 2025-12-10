/**
 * App Bundle Entry Point
 *
 * This file serves as the entry point for the main application bundle.
 * It imports all common JavaScript modules in the correct order.
 *
 * Bundle includes:
 * - HTMX core library
 * - HTMX SSE extension
 * - UI utilities (Toast, Loading, MobileMenu, etc.)
 * - Modal system
 * - Real-time notifications
 */

// Core HTMX and extensions (must be loaded first)
import '../static/js/htmx.min.js';
import '../static/js/sse.js';

// Shared utilities
import '../static/js/ui-utils.js';

// Modal handling
import '../static/js/modal.js';

// Real-time notifications via SSE
import '../static/js/notifications-realtime.js';

console.log('✅ App bundle loaded');
