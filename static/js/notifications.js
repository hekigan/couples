// Notification System
let notificationInterval = null;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // Check if user is logged in (notification bell exists)
    if (document.getElementById('notification-btn')) {
        loadNotificationCount();
        // Poll for new notifications every 30 seconds
        notificationInterval = setInterval(loadNotificationCount, 30000);
    }
});

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

