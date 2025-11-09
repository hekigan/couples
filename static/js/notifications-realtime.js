// Real-Time Notification System using SSE
let notificationEventSource = null;
let notificationReconnectTimeout = null;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // Check if user is logged in (notification bell exists)
    if (document.getElementById('notification-btn')) {
        loadNotificationCount();
        connectNotificationStream();
    }
});

// Connect to SSE notification stream
function connectNotificationStream() {
    const userId = getUserIdFromSession();
    if (!userId) {
        console.log('No user ID, skipping notification stream');
        return;
    }

    // Close existing connection if any
    if (notificationEventSource) {
        notificationEventSource.close();
    }

    console.log('Connecting to notification stream...');
    notificationEventSource = new EventSource(`/api/notifications/stream`);

    notificationEventSource.addEventListener('notification', (event) => {
        try {
            const notification = JSON.parse(event.data);
            console.log('New notification received:', notification);
            
            // Update badge count
            loadNotificationCount();
            
            // Show browser notification if permission granted
            showBrowserNotification(notification);
            
            // If dropdown is open, reload the list
            const dropdown = document.getElementById('notification-dropdown');
            if (dropdown && dropdown.style.display === 'block') {
                loadNotifications();
            }
            
            // Show toast notification
            showToastNotification(notification);
        } catch (error) {
            console.error('Error processing notification:', error);
        }
    });

    notificationEventSource.addEventListener('ping', () => {
        console.log('Notification stream ping received');
    });

    notificationEventSource.onerror = (error) => {
        console.error('Notification stream error:', error);
        notificationEventSource.close();
        
        // Reconnect after 5 seconds
        if (notificationReconnectTimeout) {
            clearTimeout(notificationReconnectTimeout);
        }
        notificationReconnectTimeout = setTimeout(() => {
            console.log('Reconnecting to notification stream...');
            connectNotificationStream();
        }, 5000);
    };

    notificationEventSource.onopen = () => {
        console.log('✅ Notification stream connected');
    };
}

// Get user ID from session (you might need to adjust this based on your auth implementation)
function getUserIdFromSession() {
    // This is a placeholder - adjust based on how you store the user session
    // Option 1: From a meta tag in the HTML
    const userMeta = document.querySelector('meta[name="user-id"]');
    if (userMeta) return userMeta.content;
    
    // Option 2: From a data attribute on the notification button
    const btn = document.getElementById('notification-btn');
    if (btn) return btn.dataset.userId;
    
    // Option 3: Assume authenticated if notification bell exists
    return 'authenticated';
}

// Show browser notification (with permission)
function showBrowserNotification(notification) {
    if (!('Notification' in window)) return;
    
    if (Notification.permission === 'granted') {
        new Notification(notification.title, {
            body: notification.message,
            icon: '/static/favicon.svg',
            badge: '/static/favicon.svg',
            tag: notification.id
        });
    } else if (Notification.permission === 'default') {
        // Request permission on first notification
        Notification.requestPermission();
    }
}

// Show toast notification in the page
function showToastNotification(notification) {
    const toast = document.createElement('div');
    toast.className = 'notification-toast';
    toast.innerHTML = `
        <div class="notification-toast-icon">${getNotificationIcon(notification.type)}</div>
        <div class="notification-toast-content">
            <div class="notification-toast-title">${notification.title}</div>
            ${notification.message ? `<div class="notification-toast-message">${notification.message}</div>` : ''}
        </div>
        <button class="notification-toast-close" onclick="this.parentElement.remove()">✕</button>
    `;
    
    document.body.appendChild(toast);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        toast.style.animation = 'slideOutRight 0.3s ease-out';
        setTimeout(() => toast.remove(), 300);
    }, 5000);
    
    // Click to navigate
    toast.addEventListener('click', (e) => {
        if (e.target.className !== 'notification-toast-close') {
            if (notification.link) {
                window.location.href = notification.link;
            }
        }
    });
}

// Load notification count
async function loadNotificationCount() {
    try {
        const response = await fetch('/api/notifications/unread-count');
        if (response.ok) {
            const data = await response.json();
            updateNotificationBadge(data.count);
        }
    } catch (error) {
        console.error('Failed to load notification count:', error);
    }
}

// Update notification badge
function updateNotificationBadge(count) {
    const badge = document.getElementById('notification-badge');
    if (count > 0) {
        badge.textContent = count > 99 ? '99+' : count;
        badge.style.display = 'flex';
    } else {
        badge.style.display = 'none';
    }
}

// Toggle notifications dropdown
async function toggleNotifications() {
    const dropdown = document.getElementById('notification-dropdown');
    const isVisible = dropdown.style.display === 'block';
    
    if (isVisible) {
        dropdown.style.display = 'none';
    } else {
        dropdown.style.display = 'block';
        await loadNotifications();
    }
}

// Close dropdown when clicking outside
document.addEventListener('click', (e) => {
    const container = document.querySelector('.notification-container');
    const dropdown = document.getElementById('notification-dropdown');
    
    if (container && !container.contains(e.target) && dropdown) {
        dropdown.style.display = 'none';
    }
});

// Load notifications
async function loadNotifications() {
    const list = document.getElementById('notification-list');
    list.innerHTML = '<p class="loading">Loading...</p>';
    
    try {
        const response = await fetch('/api/notifications');
        if (response.ok) {
            const notifications = await response.json();
            displayNotifications(notifications);
        } else {
            list.innerHTML = '<p class="error">Failed to load notifications</p>';
        }
    } catch (error) {
        console.error('Failed to load notifications:', error);
        list.innerHTML = '<p class="error">Failed to load notifications</p>';
    }
}

// Display notifications
function displayNotifications(notifications) {
    const list = document.getElementById('notification-list');
    
    if (!notifications || notifications.length === 0) {
        list.innerHTML = '<p class="empty">No notifications</p>';
        return;
    }
    
    const html = notifications.map(notification => `
        <div class="notification-item ${notification.read ? 'read' : 'unread'}" 
             onclick="handleNotificationClick('${notification.id}', '${notification.link || ''}')">
            <div class="notification-icon">
                ${getNotificationIcon(notification.type)}
            </div>
            <div class="notification-content">
                <div class="notification-title">${notification.title}</div>
                ${notification.message ? `<div class="notification-message">${notification.message}</div>` : ''}
                <div class="notification-time">${formatTime(notification.created_at)}</div>
            </div>
            ${!notification.read ? '<div class="unread-dot"></div>' : ''}
        </div>
    `).join('');
    
    list.innerHTML = html;
}

// Get icon for notification type
function getNotificationIcon(type) {
    const icons = {
        'room_invitation': '🎮',
        'friend_request': '👥',
        'game_start': '🚀',
        'message': '💬'
    };
    return icons[type] || '🔔';
}

// Format time
function formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
}

// Handle notification click
async function handleNotificationClick(notificationId, link) {
    // Mark as read
    try {
        await fetch(`/api/notifications/${notificationId}/read`, {
            method: 'POST'
        });
        
        // Update UI
        await loadNotificationCount();
        await loadNotifications();
        
        // Navigate if link provided
        if (link) {
            window.location.href = link;
        }
    } catch (error) {
        console.error('Failed to mark notification as read:', error);
    }
}

// Mark all as read
async function markAllRead() {
    try {
        const response = await fetch('/api/notifications/read-all', {
            method: 'POST'
        });
        
        if (response.ok) {
            await loadNotificationCount();
            await loadNotifications();
        }
    } catch (error) {
        console.error('Failed to mark all as read:', error);
    }
}

// Clean up on page unload
window.addEventListener('beforeunload', () => {
    if (notificationEventSource) {
        notificationEventSource.close();
    }
    if (notificationReconnectTimeout) {
        clearTimeout(notificationReconnectTimeout);
    }
});



