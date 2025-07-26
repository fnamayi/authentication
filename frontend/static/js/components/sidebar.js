// Sidebar Component for Online Users
window.SidebarComponent = {
    onlineUsers: [],
    updateInterval: null,
    eventsBound: false,

    init() {
        console.log('👥 Initializing sidebar component...');
        this.bindEvents();
        this.startUpdating();
        console.log('👥 Sidebar component initialized');
    },

    bindEvents() {
        console.log('👥 Binding sidebar events...');

        // Listen for user status updates via custom events
        window.addEventListener('user_status', (event) => {
            console.log('👥 Received user_status event:', event.detail);
            this.updateUserStatus(event.detail);
        });

        // Also try to bind directly to WebSocket if available
        if (window.forumApp?.websocket) {
            console.log('👥 WebSocket found, events will be handled via main app');
        } else {
            console.log('👥 WebSocket not found, will retry later');
            // Retry binding events after a delay if WebSocket isn't ready yet
            setTimeout(() => {
                if (window.forumApp?.websocket && !this.eventsBound) {
                    console.log('👥 WebSocket now available');
                    this.eventsBound = true;
                }
            }, 2000);
        }
    },

    startUpdating() {
        console.log('👥 Starting online users updates...');
        // Only update if user is authenticated
        const isAuthenticated = (window.auth && window.auth.isLoggedIn()) ||
                               (window.forumApp && window.forumApp.isAuthenticated);

        if (isAuthenticated) {
            console.log('👥 User is authenticated, starting updates');
            this.updateOnlineUsers();
            this.updateInterval = setInterval(() => {
                const stillAuthenticated = (window.auth && window.auth.isLoggedIn()) ||
                                         (window.forumApp && window.forumApp.isAuthenticated);
                if (stillAuthenticated) {
                    console.log('👥 Periodic update of online users');
                    this.updateOnlineUsers();
                }
            }, 30000);
        } else {
            console.log('👥 User not authenticated, skipping updates');
        }
    },

    // Method to restart updates when authentication state changes
    restart() {
        console.log('👥 Restarting sidebar component...');
        this.stopUpdating();
        this.startUpdating();
    },

    stopUpdating() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
            this.updateInterval = null;
        }
    },

    async updateOnlineUsers() {
        try {
            console.log('👥 Fetching online users from API...');
            const response = await window.api.getOnlineUsers();
            console.log('👥 API response:', response);
            if (response.success) {
                let users = response.data || [];

                // Filter out current user
                const currentUserId = window.forumApp.currentUser?.id;
                if (currentUserId) {
                    users = users.filter(user => user.userId !== currentUserId);
                    console.log('👥 Filtered out current user, remaining users:', users.length);
                }

                this.onlineUsers = users;
                console.log('👥 Online users updated:', this.onlineUsers);
                this.render();
            } else {
                console.log('👥 API response not successful:', response);
            }
        } catch (error) {
            console.error('👥 Failed to fetch online users:', error);
        }
    },

    updateUserStatus(statusData) {
        const { userId, nickname, status } = statusData;

        // Don't add/remove current user
        const currentUserId = window.forumApp.currentUser?.id;
        if (userId === currentUserId) {
            return;
        }

        if (status === 'online') {
            // Add user if not already in list
            if (!this.onlineUsers.find(user => user.userId === userId)) {
                this.onlineUsers.push({
                    userId,
                    nickname,
                    firstName: '',
                    lastName: '',
                    avatarUrl: null,
                    lastSeen: new Date().toISOString()
                });
            }
        } else if (status === 'offline') {
            // Remove user from list
            this.onlineUsers = this.onlineUsers.filter(user => user.userId !== userId);
        }

        this.render();
    },

    render() {
        console.log('👥 Rendering online users...', this.onlineUsers);
        const onlineUsersContainer = document.getElementById('online-users');
        const onlineCountElement = document.getElementById('online-count');

        if (!onlineUsersContainer || !onlineCountElement) {
            console.log('👥 Missing DOM elements:', { onlineUsersContainer, onlineCountElement });
            return;
        }

        // Update count
        onlineCountElement.textContent = this.onlineUsers.length;
        console.log('👥 Updated online count to:', this.onlineUsers.length);

        // Sort users alphabetically by nickname as specified in prompt
        const sortedUsers = [...this.onlineUsers].sort((a, b) => {
            return a.nickname.localeCompare(b.nickname);
        });

        // Render users list
        onlineUsersContainer.innerHTML = '';
        
        if (sortedUsers.length === 0) {
            onlineUsersContainer.innerHTML = `
                <div class="no-online-users">
                    <p>No users online</p>
                </div>
            `;
            return;
        }

        sortedUsers.forEach(user => {
            const userElement = this.createUserElement(user);
            onlineUsersContainer.appendChild(userElement);
        });
    },

    createUserElement(user) {
        const userDiv = document.createElement('div');
        userDiv.className = 'online-user';
        userDiv.setAttribute('data-user-id', user.userId);

        const avatarUrl = user.avatarUrl || '/static/images/default-avatar.svg';
        const displayName = user.nickname || `${user.firstName} ${user.lastName}`.trim() || 'Unknown User';

        userDiv.innerHTML = `
            <img src="${avatarUrl}" alt="${displayName}'s avatar" class="online-user-avatar">
            <span class="online-user-name">${window.utils.escapeHtml(displayName)}</span>
            <div class="online-indicator"></div>
        `;

        // Add click handler to start conversation
        userDiv.addEventListener('click', () => {
            console.log('🖱️ SIDEBAR: Click detected on online user!', user);
            console.log('🖱️ SIDEBAR: User object:', JSON.stringify(user, null, 2));
            this.startConversation(user);
        });

        return userDiv;
    },

    startConversation(user) {
        console.log('🚀 SIDEBAR: startConversation called with user:', user);
        console.log('🚀 SIDEBAR: Current user:', window.forumApp.currentUser);

        // Don't start conversation with self (this should already be filtered out, but double-check)
        if (window.forumApp.currentUser && user.userId === window.forumApp.currentUser.id) {
            console.log('🚀 SIDEBAR: Skipping conversation with self');
            return;
        }

        console.log('👥 Starting conversation with user:', user);

        // Navigate to messages page
        if (window.forumApp.router) {
            window.forumApp.router.navigate('/messages');

            // Wait for messages page to load, then select the user
            setTimeout(() => {
                if (window.MessagesPage) {
                    // Convert user object to match expected format
                    const userForMessages = {
                        id: user.userId,
                        nickname: user.nickname,
                        firstName: user.firstName || '',
                        lastName: user.lastName || '',
                        avatarURL: user.avatarUrl
                    };

                    // Start new conversation with this user
                    window.MessagesPage.selectUser(userForMessages);

                    console.log('✅ Navigated to messages and selected user:', user.nickname);
                } else {
                    console.error('❌ Messages page not available');
                }
            }, 200); // Increased timeout to ensure page loads
        }
    },

    destroy() {
        this.stopUpdating();
    }
};
