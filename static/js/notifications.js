/**
 * Notification System
 * Handles real-time notifications via SSE with polling fallback
 */

// =============================================================================
// State Management
// =============================================================================

let notificationEventSource = null;
let userEventSource = null;
let notificationReconnectTimeout = null;
let notificationsLoaded = false;
let latestNotificationTimestamp = null; // Track most recent notification to avoid duplicate alerts

// =============================================================================
// Configuration
// =============================================================================

const API_BASE = '/api/v1';

const NOTIFICATION_ICONS = {
    'room_invitation': 'icon-join-room',
    'friend_request': 'icon-people_alt',
    'friend_accepted': 'icon-sentiment_very_satisfied',
    'game_start': 'icon-sports_esports',
    'message': 'icon-comments'
};

const SSE_RECONNECT_DELAY = 5000;
const TOAST_DURATION = 5000;

// =============================================================================
// Initialization
// =============================================================================

document.addEventListener('DOMContentLoaded', () => {
    const notificationBtn = document.getElementById('notification-btn');
    if (!notificationBtn) return;

    // Check if notifications were already loaded server-side
    const notificationList = document.getElementById('notification-list');
    if (notificationList && notificationList.dataset.loaded === 'true') {
        notificationsLoaded = true; // Skip initial load

        // Set initial timestamp to now to prevent showing alerts for existing notifications
        // This assumes server-rendered notifications are current as of page load
        latestNotificationTimestamp = new Date();
    }

    // Initial badge count load (notifications list is already server-rendered)
    loadNotificationCount();

    // Connect to real-time streams
    connectNotificationStream();
    connectUserEventStream();

    // Lazy load notifications on hover/focus (only if not already loaded)
    const loadOnInteraction = () => {
        if (!notificationsLoaded) {
            notificationsLoaded = true;
            loadNotifications();
        }
    };
    notificationBtn.addEventListener('mouseenter', loadOnInteraction);
    notificationBtn.addEventListener('focus', loadOnInteraction);
});

// Clean up on page unload
window.addEventListener('beforeunload', () => {
    if (notificationEventSource) {
        notificationEventSource.close();
    }
    if (userEventSource) {
        userEventSource.close();
    }
    if (notificationReconnectTimeout) {
        clearTimeout(notificationReconnectTimeout);
    }
});

// =============================================================================
// SSE Stream Connections
// =============================================================================

function connectNotificationStream() {
    if (!getUserIdFromSession()) {
        console.log('No user ID, skipping notification stream');
        return;
    }

    // Close existing connection
    if (notificationEventSource) {
        notificationEventSource.close();
    }

    console.log('Connecting to notification stream...');
    notificationEventSource = new EventSource(`${API_BASE}/stream/notifications`);

    notificationEventSource.addEventListener('notification', (event) => {
        try {
            const notification = JSON.parse(event.data);
            console.log('New notification received:', notification);

            // Update badge count
            loadNotificationCount();

            // Only show browser/toast notification if it's newer than what's already displayed
            const notificationTime = new Date(notification.created_at);

            // Defensive check: ensure valid date was parsed
            if (isNaN(notificationTime.getTime())) {
                console.error('Invalid notification timestamp:', notification.created_at);
                return;
            }

            const isNewerThanDisplayed = !latestNotificationTimestamp ||
                                         notificationTime > latestNotificationTimestamp;

            if (isNewerThanDisplayed) {
                // Show browser notification
                showBrowserNotification(notification);

                // Show toast notification
                if (typeof Toast !== 'undefined') {
                    Toast.show({
                        type: 'info',
                        title: notification.title,
                        message: notification.message,
                        icon: `<i class="${getNotificationIcon(notification.type)}"></i>`,
                        duration: TOAST_DURATION
                    });
                }

                // Update latest timestamp
                latestNotificationTimestamp = notificationTime;
            }
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

        // Reconnect after delay
        if (notificationReconnectTimeout) {
            clearTimeout(notificationReconnectTimeout);
        }
        notificationReconnectTimeout = setTimeout(() => {
            console.log('Reconnecting to notification stream...');
            connectNotificationStream();
        }, SSE_RECONNECT_DELAY);
    };

    notificationEventSource.onopen = () => {
        console.log('✅ Notification stream connected');
    };
}

function connectUserEventStream() {
    if (!getUserIdFromSession()) {
        console.log('No user ID, skipping user event stream');
        return;
    }

    // Close existing connection
    if (userEventSource) {
        userEventSource.close();
    }

    console.log('Connecting to user event stream...');
    userEventSource = new EventSource(`${API_BASE}/stream/user/events`);

    userEventSource.addEventListener('badge_update', (event) => {
        try {
            const data = JSON.parse(event.data);
            console.log('Badge update received:', data);
            updateNotificationBadge(data.count);
        } catch (error) {
            console.error('Error processing badge update:', error);
        }
    });

    userEventSource.onerror = (error) => {
        console.error('User event stream error:', error);
        userEventSource.close();

        // Reconnect after delay
        setTimeout(() => {
            console.log('Reconnecting to user event stream...');
            connectUserEventStream();
        }, SSE_RECONNECT_DELAY);
    };

    userEventSource.onopen = () => {
        console.log('✅ User event stream connected');
    };
}

// =============================================================================
// API Functions
// =============================================================================

async function loadNotificationCount() {
    try {
        const response = await fetch(`${API_BASE}/notifications/unread-count`);
        const contentType = response.headers.get('content-type');

        if (response.ok && contentType && contentType.includes('application/json')) {
            const data = await response.json();
            updateNotificationBadge(data.count);
        } else if (response.status === 401 || response.status === 403) {
            console.log('Not authenticated, skipping notification count');
        }
    } catch (error) {
        console.error('Failed to load notification count:', error);
    }
}

async function loadNotifications() {
    const list = document.getElementById('notification-list');
    if (!list) return;

    list.innerHTML = '<div class="loading">Loading...</div>';

    try {
        const response = await fetch(`${API_BASE}/notifications`);
        const contentType = response.headers.get('content-type');

        if (response.ok && contentType && contentType.includes('application/json')) {
            const notifications = await response.json();
            displayNotifications(notifications);
            list.dataset.loaded = 'true'; // Mark as loaded after successful fetch
        } else if (response.status === 401 || response.status === 403) {
            list.innerHTML = '<div class="error">Please log in to view notifications</div>';
        } else {
            list.innerHTML = '<div class="error">Failed to load notifications</div>';
        }
    } catch (error) {
        console.error('Failed to load notifications:', error);
        list.innerHTML = '<div class="error">Failed to load notifications</div>';
    }
}

// Removed handleNotificationClick - notifications now use direct links

async function markAllRead() {
    try {
        const response = await fetch(`${API_BASE}/notifications/read-all`, {
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

// =============================================================================
// UI Functions
// =============================================================================

function updateNotificationBadge(count) {
    const badge = document.getElementById('notification-badge');
    if (!badge) return;

    if (count > 0) {
        badge.textContent = count > 99 ? '99+' : count;
        badge.removeAttribute('hidden');
        badge.style.display = 'inline-flex';
        badge.classList.add('visible');
    } else {
        badge.setAttribute('hidden', '');
        badge.style.display = 'none';
        badge.classList.remove('visible');
    }
}

function displayNotifications(notifications) {
    const list = document.getElementById('notification-list');
    if (!list) return;

    if (!notifications || notifications.length === 0) {
        list.innerHTML = '<div class="empty">No notifications</div>';
        return;
    }

    // Track the most recent notification timestamp
    const mostRecent = notifications.reduce((latest, notif) => {
        const notifTime = new Date(notif.created_at);
        return !latest || notifTime > latest ? notifTime : latest;
    }, null);

    if (mostRecent) {
        latestNotificationTimestamp = mostRecent;
    }

    // Aggregate notifications by type
    const aggregated = aggregateNotificationsByType(notifications);

    const html = aggregated.map(agg => `
        <a href="${agg.link}" class="notification-item ${agg.hasUnread ? 'unread' : ''}">
            <div class="notification-icon">
                <i class="${agg.icon}"></i>
            </div>
            <div class="notification-content">
                <div class="notification-title">
                    <span class="notification-count-badge">${agg.count}</span>
                    ${pluralizeLabel(agg.label, agg.count)}
                </div>
            </div>
            ${agg.hasUnread ? '<div class="unread-dot"></div>' : ''}
        </a>
    `).join('');

    list.innerHTML = html;
}

function aggregateNotificationsByType(notifications) {
    const typeMap = {};

    const typeLabels = {
        'room_invitation': 'Room invitation',
        'friend_request': 'Friend request',
        'friend_accepted': 'Friend accepted',
        'game_start': 'Game started',
        'message': 'Message'
    };

    const typeLinks = {
        'room_invitation': '/game/rooms',
        'friend_request': '/friends',
        'friend_accepted': '/friends',
        'game_start': '/game/rooms',
        'message': '/'
    };

    notifications.forEach(notif => {
        if (typeMap[notif.type]) {
            typeMap[notif.type].count++;
            if (!notif.read) {
                typeMap[notif.type].hasUnread = true;
            }
        } else {
            typeMap[notif.type] = {
                type: notif.type,
                count: 1,
                icon: getNotificationIcon(notif.type),
                label: typeLabels[notif.type] || 'Notification',
                hasUnread: !notif.read,
                link: typeLinks[notif.type] || '/'
            };
        }
    });

    return Object.values(typeMap);
}

function pluralizeLabel(label, count) {
    return count > 1 ? label + '(s)' : label;
}

// =============================================================================
// Browser Notifications
// =============================================================================

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
        Notification.requestPermission();
    }
}

// =============================================================================
// Utility Functions
// =============================================================================

function getUserIdFromSession() {
    // Option 1: From a meta tag in the HTML
    const userMeta = document.querySelector('meta[name="user-id"]');
    if (userMeta) return userMeta.content;

    // Option 2: From a data attribute on the notification button
    const btn = document.getElementById('notification-btn');
    if (btn) return btn.dataset.userId;

    // Option 3: Assume authenticated if notification bell exists
    return 'authenticated';
}

function getNotificationIcon(type) {
    return NOTIFICATION_ICONS[type] || '';
}

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
