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

	const url = '/admin/api/' + type + '/bulk-' + action;

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
 * Show a modal dialog using native HTMLDialogElement API
 * @param {string} modalId - The ID of the dialog to show
 */
function showModal(modalId) {
	const dialog = document.getElementById(modalId);
	if (dialog && dialog.tagName === 'DIALOG') {
		// Add animation class
		document.documentElement.classList.add('modal-is-opening');

		// Show modal
		dialog.showModal();

		// Remove animation class after animation completes
		setTimeout(() => {
			document.documentElement.classList.remove('modal-is-opening');
		}, 400);
	}
}

/**
 * Hide a modal dialog using native HTMLDialogElement API
 * @param {string} modalId - The ID of the dialog to hide
 */
function hideModal(modalId) {
	const dialog = document.getElementById(modalId);
	if (dialog && dialog.tagName === 'DIALOG') {
		// Add closing animation
		document.documentElement.classList.add('modal-is-closing');

		// Close after animation
		setTimeout(() => {
			dialog.close();
			document.documentElement.classList.remove('modal-is-closing');
		}, 400);
	}
}

/**
 * Close dialog when clicking on backdrop (outside dialog content)
 */
document.addEventListener('click', function(event) {
	if (event.target.tagName === 'DIALOG') {
		const rect = event.target.getBoundingClientRect();
		const isInDialog = (
			rect.top <= event.clientY &&
			event.clientY <= rect.top + rect.height &&
			rect.left <= event.clientX &&
			event.clientX <= rect.left + rect.width
		);

		if (!isInDialog) {
			return;
		}

		// Check if click is outside the article (modal content)
		if (event.target === event.currentTarget) {
			hideModal(event.target.id);
		}
	}
});

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
