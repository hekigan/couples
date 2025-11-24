/*
 * Modal
 *
 * Pico.css - https://picocss.com
 * Copyright 2019-2024 - Licensed under MIT
 */

// Config
const isOpenClass = "modal-is-open";
const openingClass = "modal-is-opening";
const closingClass = "modal-is-closing";
const scrollbarWidthCssVar = "--pico-scrollbar-width";
const animationDuration = 400; // ms
let visibleModal = null;

// Toggle modal
const toggleModal = (event) => {
  event.preventDefault();
  const modal = document.getElementById(event.currentTarget.dataset.target);
  if (!modal) return;
  modal && (modal.open ? closeModal(modal) : openModal(modal));
};

// Open modal
const openModal = (modal) => {
  const { documentElement: html } = document;
  const scrollbarWidth = getScrollbarWidth();
  if (scrollbarWidth) {
    html.style.setProperty(scrollbarWidthCssVar, `${scrollbarWidth}px`);
  }
  html.classList.add(isOpenClass, openingClass);
  setTimeout(() => {
    visibleModal = modal;
    html.classList.remove(openingClass);
  }, animationDuration);
  modal.showModal();
};

// Close modal
const closeModal = (modal) => {
  visibleModal = null;
  const { documentElement: html } = document;
  html.classList.add(closingClass);
  setTimeout(() => {
    html.classList.remove(closingClass, isOpenClass);
    html.style.removeProperty(scrollbarWidthCssVar);
    modal.close();
  }, animationDuration);
};

// Close with a click outside
document.addEventListener("click", (event) => {
  if (visibleModal === null) return;
  const modalContent = visibleModal.querySelector("article");
  const isClickInside = modalContent.contains(event.target);
  !isClickInside && closeModal(visibleModal);
});

// Close with Esc key
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && visibleModal) {
    closeModal(visibleModal);
  }
});

// Get scrollbar width
const getScrollbarWidth = () => {
  const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
  return scrollbarWidth;
};

// Is scrollbar visible
const isScrollbarVisible = () => {
  return document.body.scrollHeight > screen.height;
};

// Submit modal form
function submitModalForm(event) {
    const modal = event.currentTarget.closest('dialog');
    const form = modal.querySelector('.modal-content form');
    
    if (!form) {
      alert('No modal content form found');
        return;
    }
    
    // Trigger form submit (HTMX will intercept if form has hx-* attributes)
    form.requestSubmit();
}

function handleDataUpdateResponse(event, apiPath = '/admin/api/questions/list?page=1', targetSelector = '#questions-list') {
	const xhr = event.detail.xhr;

	if (xhr.status === 200) {
		// Success: show toast and close modal
		try {
			const response = JSON.parse(xhr.responseText);
			showToast(response.success || 'Data updated successfully', 'success');
		} catch (e) {
			showToast('Data updated successfully', 'success');
		}

		// Close the modal
		const modal = event.currentTarget.closest('dialog');
		if (modal) {
			modal.close();
		}

		// Refresh the questions list
        apiPath = apiPath + window.location.search;
		htmx.ajax('GET', apiPath, {
			target: targetSelector,
			swap: 'outerHTML',
			indicator: '#questions-list-loading'
		});
	} else {
		// Error: show error toast
		try {
			const response = JSON.parse(xhr.responseText);
			showToast(response.error || 'Failed to update data', 'error');
		} catch (e) {
			showToast('Failed to update data', 'error');
		}
	}
}