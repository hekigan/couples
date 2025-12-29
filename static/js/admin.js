/**
 * Admin Panel JavaScript
 * Handles bulk actions and interactive features for admin pages
 */

/**
 * Submit bulk action for selected items
 * @param {string} type - The type of items (users, questions, categories, rooms)
 */
function submitBulkAction(type) {
	const form = document.getElementById('bulk-' + type + '-form');
	const action = document.getElementById('bulk-action-' + type).value || form.querySelector('[name="bulk_action_bottom"]').value;

	if (!action) {
		alert('Please select an action');
		return;
	}

	const checkboxes = form.querySelectorAll('input[type="checkbox"]:checked:not([id^="select-all"])');
	if (checkboxes.length === 0) {
		alert('Please select at least one item');
		return;
	}

	const actionName = action === 'delete' ? 'delete' : action === 'close' ? 'close' : action;
	if (!confirm('Are you sure you want to ' + actionName + ' ' + checkboxes.length + ' selected item(s)?')) {
		return;
	}

	const url = '/admin/api/v1/' + type + '/bulk-' + action;

	// Create URL-encoded form data to handle arrays correctly
	const formData = new FormData(form);
	const urlEncoded = new URLSearchParams(formData).toString();

	fetch(url, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		body: urlEncoded
	})
	.then(response => response.text())
	.then(html => {
		document.getElementById(type + '-list').outerHTML = html;
	})
	.catch(error => {
		console.error('Error:', error);
		alert('Failed to perform bulk action');
	});
}

/**
 * Toggle all checkboxes in a form
 * @param {HTMLElement} checkbox - The "select all" checkbox element
 * @param {string} targetClass - The class name of checkboxes to toggle
 */
function toggleAllCheckboxes(checkbox, targetClass) {
	const checkboxes = document.querySelectorAll('.' + targetClass);
	checkboxes.forEach(cb => {
		cb.checked = checkbox.checked;
	});
}

/**
 * Save items-per-page preference to localStorage
 * Used across all admin tables for consistent pagination
 * @param {string|number} value - The number of items per page (25, 50, or 100)
 */
function savePerPagePref(value) {
	try {
		localStorage.setItem('admin_table_per_page', value.toString());
		console.log('Saved per-page preference:', value);
	} catch (e) {
		console.error('Failed to save per-page preference:', e);
	}
}

/**
 * Get items-per-page preference from localStorage
 * @returns {number} The saved preference or default (25)
 */
function getPerPagePref() {
	try {
		const saved = localStorage.getItem('admin_table_per_page');
		if (saved) {
			const value = parseInt(saved, 10);
			// Validate against allowed values
			if (value === 25 || value === 50 || value === 100) {
				return value;
			}
		}
	} catch (e) {
		console.error('Failed to get per-page preference:', e);
	}
	return 25; // Default
}

/**
 * Initialize per-page selects with saved preference
 * Called on page load for admin pages
 */
document.addEventListener('DOMContentLoaded', function() {
	const perPageSelect = document.getElementById('per-page-select');
	if (perPageSelect) {
		const savedValue = getPerPagePref();
		perPageSelect.value = savedValue.toString();
	}
});

// Note: showToast() is now available from ui-utils.js
// It uses the shared Toast notification system
// Example usage: showToast('Operation successful', 'success')

// ============================================================================
// Routes Page - Client-Side Filtering
// ============================================================================

function filterRoutes() {
	const searchTerm = document.getElementById('route-search').value.toLowerCase();
	const filterGET = document.getElementById('filter-get').checked;
	const filterPOST = document.getElementById('filter-post').checked;
	const filterPUT = document.getElementById('filter-put').checked;
	const filterDELETE = document.getElementById('filter-delete').checked;
	const htmxOnly = document.getElementById('filter-htmx').checked;

	const rows = document.querySelectorAll('.route-row');
	let visibleCount = 0;

	rows.forEach(row => {
		const method = row.dataset.method;
		const path = row.dataset.path.toLowerCase();
		const isHTMX = row.dataset.isHtmx === 'true';

		// Method filter
		let methodMatch = false;
		if (method === 'GET' && filterGET) methodMatch = true;
		if (method === 'POST' && filterPOST) methodMatch = true;
		if (method === 'PUT' && filterPUT) methodMatch = true;
		if (method === 'DELETE' && filterDELETE) methodMatch = true;

		// HTMX filter
		const htmxMatch = !htmxOnly || isHTMX;

		// Search term (matches path or method)
		const searchMatch = !searchTerm || path.includes(searchTerm) || method.toLowerCase().includes(searchTerm);

		// Show/hide row
		const visible = methodMatch && htmxMatch && searchMatch;
		row.style.display = visible ? '' : 'none';

		if (visible) visibleCount++;
	});

	// Update count
	document.getElementById('visible-count').textContent = visibleCount;

	// Hide empty groups
	document.querySelectorAll('.route-group').forEach(group => {
		const visibleRows = group.querySelectorAll('.route-row:not([style*="display: none"])').length;
		const groupCount = group.querySelector('.route-group-count');

		group.style.display = visibleRows === 0 ? 'none' : '';
		if (visibleRows > 0) {
			groupCount.textContent = `(${visibleRows})`;
		}
	});
}

function toggleRouteGroup(button) {
	const group = button.closest('.route-group');
	const content = group.querySelector('.route-group-content');
	const icon = button.querySelector('.toggle-icon');
	const isExpanded = button.getAttribute('aria-expanded') === 'true';

	content.style.display = isExpanded ? 'none' : '';
	button.setAttribute('aria-expanded', !isExpanded);
	icon.textContent = isExpanded ? '▶' : '▼';
}
