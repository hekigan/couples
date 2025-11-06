/**
 * UI Utilities - Toast Notifications, Loading States, and Animations
 */

// Toast Notification System
const Toast = {
    container: null,

    init() {
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.className = 'toast-container';
            document.body.appendChild(this.container);
        }
    },

    show(options) {
        this.init();

        const {
            type = 'info',
            title = '',
            message = '',
            duration = 3000,
            icon = null
        } = options;

        const toast = document.createElement('div');
        toast.className = `toast toast-${type} fade-in`;

        const icons = {
            success: '✓',
            error: '✗',
            warning: '⚠',
            info: 'ℹ'
        };

        const toastIcon = icon || icons[type] || icons.info;

        toast.innerHTML = `
            <div class="toast-icon">${toastIcon}</div>
            <div class="toast-content">
                ${title ? `<div class="toast-title">${title}</div>` : ''}
                ${message ? `<div class="toast-message">${message}</div>` : ''}
            </div>
            <button class="toast-close" onclick="Toast.close(this.parentElement)">×</button>
            ${duration > 0 ? '<div class="toast-progress"></div>' : ''}
        `;

        this.container.appendChild(toast);

        // Auto-remove after duration
        if (duration > 0) {
            setTimeout(() => {
                this.close(toast);
            }, duration);
        }

        return toast;
    },

    close(toast) {
        if (!toast) return;
        toast.style.animation = 'slideOutRight 0.3s ease-out';
        setTimeout(() => {
            if (toast.parentElement) {
                toast.parentElement.removeChild(toast);
            }
        }, 300);
    },

    success(message, title = 'Success') {
        return this.show({ type: 'success', title, message });
    },

    error(message, title = 'Error') {
        return this.show({ type: 'error', title, message });
    },

    warning(message, title = 'Warning') {
        return this.show({ type: 'warning', title, message });
    },

    info(message, title = 'Info') {
        return this.show({ type: 'info', title, message });
    }
};

// Loading Overlay
const Loading = {
    overlay: null,

    show(message = 'Loading...') {
        if (this.overlay) return;

        this.overlay = document.createElement('div');
        this.overlay.className = 'loading-overlay';
        this.overlay.innerHTML = `
            <div class="loading-overlay-content">
                <div class="loading-spinner loading-spinner-large"></div>
                <p style="margin-top: 20px; font-size: 16px; color: #4b5563;">${message}</p>
            </div>
        `;

        document.body.appendChild(this.overlay);
        document.body.style.overflow = 'hidden';
    },

    hide() {
        if (!this.overlay) return;

        this.overlay.style.animation = 'fadeOut 0.3s ease-out';
        setTimeout(() => {
            if (this.overlay && this.overlay.parentElement) {
                this.overlay.parentElement.removeChild(this.overlay);
                this.overlay = null;
                document.body.style.overflow = '';
            }
        }, 300);
    },

    isShowing() {
        return this.overlay !== null;
    }
};

// Button Loading State
function setButtonLoading(button, loading = true, originalText = null) {
    if (loading) {
        button.dataset.originalText = button.textContent;
        button.disabled = true;
        button.innerHTML = `
            <span class="loading-spinner loading-spinner-small" style="margin-right: 8px;"></span>
            ${originalText || 'Loading...'}
        `;
    } else {
        button.disabled = false;
        button.textContent = button.dataset.originalText || originalText || 'Submit';
        delete button.dataset.originalText;
    }
}

// Animate Element
function animateElement(element, animation = 'bounce-in') {
    element.classList.add(animation);
    element.addEventListener('animationend', () => {
        element.classList.remove(animation);
    }, { once: true });
}

// Shake Element (for errors)
function shakeElement(element) {
    element.classList.add('shake');
    element.addEventListener('animationend', () => {
        element.classList.remove('shake');
    }, { once: true });
}

// Smooth Scroll to Element
function smoothScrollTo(element) {
    element.scrollIntoView({
        behavior: 'smooth',
        block: 'center'
    });
}

// Debounce Function
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Confirm Dialog with Better UX
function confirmAction(message, onConfirm, onCancel = null) {
    const confirmed = confirm(message);
    if (confirmed) {
        if (onConfirm) onConfirm();
    } else {
        if (onCancel) onCancel();
    }
    return confirmed;
}

// Copy to Clipboard with Toast
async function copyToClipboard(text, successMessage = 'Copied to clipboard!') {
    try {
        await navigator.clipboard.writeText(text);
        Toast.success(successMessage);
        return true;
    } catch (err) {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            Toast.success(successMessage);
            return true;
        } catch (err) {
            Toast.error('Failed to copy');
            return false;
        } finally {
            document.body.removeChild(textArea);
        }
    }
}

// Network Request with Loading and Error Handling
async function fetchWithUI(url, options = {}, showLoading = true) {
    if (showLoading) {
        Loading.show('Processing...');
    }

    try {
        const response = await fetch(url, options);

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `HTTP ${response.status}`);
        }

        const data = await response.json();
        return { success: true, data };
    } catch (error) {
        Toast.error(error.message || 'An error occurred');
        return { success: false, error: error.message };
    } finally {
        if (showLoading) {
            Loading.hide();
        }
    }
}

// Form Validation with Visual Feedback
function validateField(input, validationFn, errorMessage) {
    const isValid = validationFn(input.value);

    const errorElement = input.parentElement.querySelector('.error-message') ||
        document.createElement('div');

    if (!isValid) {
        input.classList.add('error');
        input.classList.remove('success');
        shakeElement(input);

        errorElement.className = 'error-message';
        errorElement.style.color = '#ef4444';
        errorElement.style.fontSize = '14px';
        errorElement.style.marginTop = '4px';
        errorElement.textContent = errorMessage;

        if (!input.parentElement.querySelector('.error-message')) {
            input.parentElement.appendChild(errorElement);
        }
    } else {
        input.classList.remove('error');
        input.classList.add('success');

        if (input.parentElement.querySelector('.error-message')) {
            errorElement.remove();
        }
    }

    return isValid;
}

// Initialize UI utilities on page load
document.addEventListener('DOMContentLoaded', () => {
    Toast.init();

    // Add fade-in animation to main content
    const mainContent = document.querySelector('main') || document.querySelector('.container');
    if (mainContent) {
        animateElement(mainContent, 'fade-in');
    }
});

// Export utilities
window.Toast = Toast;
window.Loading = Loading;
window.setButtonLoading = setButtonLoading;
window.animateElement = animateElement;
window.shakeElement = shakeElement;
window.smoothScrollTo = smoothScrollTo;
window.debounce = debounce;
window.confirmAction = confirmAction;
window.copyToClipboard = copyToClipboard;
window.fetchWithUI = fetchWithUI;
window.validateField = validateField;
