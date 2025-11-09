// Real-time client for Server-Sent Events (SSE)

class RealtimeClient {
    constructor(roomId) {
        this.roomId = roomId;
        this.eventSource = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
    }

    connect() {
        const url = `/api/rooms/${this.roomId}/events`;
        console.log('Connecting to SSE:', url);

        this.eventSource = new EventSource(url);

        this.eventSource.onopen = () => {
            console.log('SSE connection established');
            this.reconnectAttempts = 0;
        };

        this.eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            this.eventSource.close();
            this.reconnect();
        };

        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleEvent(data);
            } catch (error) {
                console.error('Failed to parse SSE data:', error);
            }
        };
    }

    reconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;
        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

        setTimeout(() => this.connect(), delay);
    }

    handleEvent(event) {
        console.log('Received event:', event);

        switch (event.type) {
            case 'room_update':
                this.handleRoomUpdate(event.data);
                break;
            case 'player_joined':
                this.handlePlayerJoined(event.data);
                break;
            case 'player_left':
                this.handlePlayerLeft(event.data);
                break;
            case 'game_started':
                this.handleGameStarted(event.data);
                break;
            case 'question_drawn':
                this.handleQuestionDrawn(event.data);
                break;
            case 'answer_submitted':
                this.handleAnswerSubmitted(event.data);
                break;
            case 'turn_changed':
                this.handleTurnChanged(event.data);
                break;
            case 'game_finished':
                this.handleGameFinished(event.data);
                break;
            case 'room_deleted':
                this.handleRoomDeleted(event.data);
                break;
            default:
                console.warn('Unknown event type:', event.type);
        }
    }

    handleRoomUpdate(data) {
        console.log('Room updated:', data);
        // Refresh the page or update specific elements
        htmx.ajax('GET', window.location.href, {target: 'body', swap: 'innerHTML'});
    }

    handlePlayerJoined(data) {
        console.log('Player joined:', data);
        this.showNotification('A player has joined the room', 'success');
        // Refresh player list
        htmx.ajax('GET', `/api/rooms/${this.roomId}/players`, {target: '#player-list', swap: 'innerHTML'});
    }

    handlePlayerLeft(data) {
        console.log('Player left:', data);
        this.showNotification('A player has left the room', 'info');
        // Refresh player list
        htmx.ajax('GET', `/api/rooms/${this.roomId}/players`, {target: '#player-list', swap: 'innerHTML'});
    }

    handleGameStarted(data) {
        console.log('Game started:', data);
        this.showNotification('Game is starting!', 'success');
        // Redirect to game page
        window.location.href = `/game/play/${this.roomId}`;
    }

    handleQuestionDrawn(data) {
        console.log('Question drawn:', data);
        // Update question display
        htmx.ajax('GET', `/game/play/${this.roomId}`, {target: '#game-content', swap: 'innerHTML'});
    }

    handleAnswerSubmitted(data) {
        console.log('Answer submitted:', data);
        this.showNotification('Answer submitted!', 'success');
        // Refresh game state
        htmx.ajax('GET', `/game/play/${this.roomId}`, {target: '#game-content', swap: 'innerHTML'});
    }

    handleTurnChanged(data) {
        console.log('Turn changed:', data);
        // Update turn indicator
        htmx.ajax('GET', `/game/play/${this.roomId}`, {target: '#game-content', swap: 'innerHTML'});
    }

    handleGameFinished(data) {
        console.log('Game finished:', data);
        this.showNotification('Game finished!', 'success');
        // Redirect to results page
        setTimeout(() => {
            window.location.href = `/game/finished/${this.roomId}`;
        }, 1500);
    }

    handleRoomDeleted(data) {
        console.log('Room deleted:', data);
        this.showNotification('This room has been deleted by the owner', 'error');
        this.disconnect();
        // Redirect to home after 3 seconds
        setTimeout(() => {
            window.location.href = '/';
        }, 3000);
    }

    showNotification(message, type = 'info') {
        // Simple notification system
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : '#2196F3'};
            color: white;
            border-radius: 4px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.2);
            z-index: 10000;
            animation: slideIn 0.3s ease-out;
        `;

        document.body.appendChild(notification);

        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s ease-out';
            setTimeout(() => notification.remove(), 300);
        }, 3000);
    }

    disconnect() {
        if (this.eventSource) {
            console.log('Disconnecting SSE');
            this.eventSource.close();
            this.eventSource = null;
        }
    }
}

// Auto-initialize if room ID is present
document.addEventListener('DOMContentLoaded', () => {
    const roomElement = document.querySelector('[data-room-id]');
    if (roomElement) {
        const roomId = roomElement.dataset.roomId;
        window.realtimeClient = new RealtimeClient(roomId);
        window.realtimeClient.connect();
    }
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (window.realtimeClient) {
        window.realtimeClient.disconnect();
    }
});



