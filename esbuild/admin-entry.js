/**
 * Admin Bundle Entry Point
 *
 * This file serves as the entry point for the admin-specific bundle.
 * It imports admin panel utilities.
 *
 * Bundle includes:
 * - Admin panel utilities (bulk actions, pagination, etc.)
 *
 * Note: This bundle depends on app.bundle.js being loaded first,
 * as it uses the showToast() function from ui-utils.js.
 */

// Admin panel utilities
import '../static/js/admin.js';

console.log('✅ Admin bundle loaded');
