// Enhanced Notification Component
window.NotificationComponent = {
    notifications: [],
    container: null,

    init() {
        this.createContainer();
        this.bindEvents();
        console.log('🔔 Notification component initialized');
    },

    createContainer() {
        if (this.container) return;
        
        this.container = document.createElement('div');
        this.container.id = 'notification-container';
        this.container.className = 'notification-container';
        document.body.appendChild(this.container);
    },

    show(message, type = 'info', duration = 5000, options = {}) {
        const notification = {
            id: Date.now() + Math.random(),
            message,
            type,
            duration,
            timestamp: new Date(),
            ...options
        };

        this.notifications.push(notification);
        this.render(notification);

        if (duration > 0) {
            setTimeout(() => {
                this.remove(notification.id);
            }, duration);
        }

        return notification.id;
    },

    render(notification) {
        const element = document.createElement('div');
        element.className = `notification notification-${notification.type} notification-enter`;
        element.setAttribute('data-id', notification.id);
        
        const icon = this.getIcon(notification.type);
        
        element.innerHTML = `
            <div class="notification-content">
                <div class="notification-icon">
                    ${icon}
                </div>
                <div class="notification-body">
                    <div class="notification-message">${window.utils.escapeHtml(notification.message)}</div>
                    ${notification.title ? `<div class="notification-title">${window.utils.escapeHtml(notification.title)}</div>` : ''}
                </div>
                <button class="notification-close" onclick="window.forumApp.notificationComponent.remove(${notification.id})">
                    ×
                </button>
            </div>
            <div class="notification-progress"></div>
        `;

        this.container.appendChild(element);

        // Trigger animation
        setTimeout(() => {
            element.classList.remove('notification-enter');
            element.classList.add('notification-visible');
        }, 10);

        // Add progress bar animation
        if (notification.duration > 0) {
            const progressBar = element.querySelector('.notification-progress');
            progressBar.style.animationDuration = `${notification.duration}ms`;
            progressBar.classList.add('notification-progress-active');
        }
    },

    getIcon(type) {
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️',
            loading: '⏳'
        };
        return icons[type] || icons.info;
    },

    remove(id) {
        const element = this.container.querySelector(`[data-id="${id}"]`);
        if (!element) return;

        element.classList.remove('notification-visible');
        element.classList.add('notification-exit');

        setTimeout(() => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        }, 300);

        this.notifications = this.notifications.filter(n => n.id !== id);
    },

    clear() {
        this.notifications.forEach(notification => {
            this.remove(notification.id);
        });
    },

    // Convenience methods
    success(message, duration = 5000, options = {}) {
        return this.show(message, 'success', duration, options);
    },

    error(message, duration = 8000, options = {}) {
        return this.show(message, 'error', duration, options);
    },

    warning(message, duration = 6000, options = {}) {
        return this.show(message, 'warning', duration, options);
    },

    info(message, duration = 5000, options = {}) {
        return this.show(message, 'info', duration, options);
    },

    loading(message, options = {}) {
        return this.show(message, 'loading', 0, options);
    },

    bindEvents() {
        // Listen for WebSocket notifications
        if (window.forumApp?.websocket) {
            window.forumApp.websocket.addEventListener('notification', (event) => {
                const notification = event.detail;
                this.show(notification.message, notification.type || 'info');
            });
        }
    }
};
