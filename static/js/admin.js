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
