class MessagesPage {
    constructor() {
        this.currentUser = null;
        this.conversations = [];
        this.onlineUsers = [];
        this.currentChatUser = null;
        this.currentChatWindow = null;
        this.messageSound = null;
        this.initialized = false;

        // Pagination tracking for each chat
        this.chatPagination = new Map(); // userId -> { offset, hasMore, loading }
        this.scrollThrottleTimeout = null;
    }

    async init() {
        if (this.initialized) {
            console.log('🔄 Messages Page already initialized, refreshing data...');
            await this.refreshData();
            return;
        }

        console.log('🔄 Initializing Messages Page...');

        // Get current user from main app instead of making another API call
        if (window.forumApp && window.forumApp.currentUser) {
            this.currentUser = window.forumApp.currentUser;
            console.log('✅ Current user loaded from main app:', this.currentUser.nickname);
        } else {
            // Fallback to API call if main app user not available
            try {
                const response = await fetch('/api/user');
                const data = await response.json();
                if (data.success) {
                    this.currentUser = data.data;
                    console.log('✅ Current user loaded from API:', this.currentUser.nickname);
                } else {
                    console.error('❌ Failed to get current user');
                    return;
                }
            } catch (error) {
                console.error('❌ Error getting current user:', error);
                return;
            }
        }

        // Load initial data
        await this.loadConversations();
        await this.loadOnlineUsers();

        // Setup event listeners
        this.setupEventListeners();

        // Initialize notification sound
        this.initNotificationSound();

        this.initialized = true;
        console.log('✅ Messages Page initialized');
    }

    async refreshData() {
        console.log('🔄 Refreshing messages page data...');
        await this.loadConversations();
        await this.loadOnlineUsers();
    }

    cleanup() {
        console.log('🧹 Cleaning up Messages Page...');

        // Close any open chat windows and remove scroll listeners
        if (this.currentChatWindow) {
            const container = this.currentChatWindow.querySelector('[id^="chatMessages-"]');
            if (container && container.scrollHandler) {
                container.removeEventListener('scroll', container.scrollHandler);
                container.scrollHandler = null;
            }
            this.currentChatWindow.remove();
            this.currentChatWindow = null;
        }

        // Reset current chat user
        this.currentChatUser = null;

        // Clear pagination data
        this.chatPagination.clear();

        // Clear throttle timeout
        if (this.scrollThrottleTimeout) {
            clearTimeout(this.scrollThrottleTimeout);
            this.scrollThrottleTimeout = null;
        }

        // Note: We don't clean up the WebSocket since we're using the main app's WebSocket
    }

    // Use main app's WebSocket instead of creating a new one
    getWebSocket() {
        return window.forumApp?.websocket;
    }

    // Throttle function to prevent scroll event spam
    throttle(func, delay) {
        return (...args) => {
            if (this.scrollThrottleTimeout) {
                clearTimeout(this.scrollThrottleTimeout);
            }
            this.scrollThrottleTimeout = setTimeout(() => func.apply(this, args), delay);
        };
    }

    handleWebSocketMessage(message) {
        console.log('📨 Messages Page: WebSocket message received:', message);

        switch (message.type) {
            case 'new_message':
                console.log('📩 Messages Page: Handling new message');
                this.handleNewMessage(message.data);
                break;
            case 'message_read':
                console.log('📖 Messages Page: Handling message read');
                this.handleMessageRead(message.data);
                break;
            case 'user_status':
                console.log('👤 Messages Page: Handling user status');
                this.handleUserStatus(message.data);
                break;
            default:
                console.log('❓ Messages Page: Unknown WebSocket message type:', message.type);
        }
    }

    handleNewMessage(messageData) {
        console.log('📩 New message received:', messageData);
        
        // Play notification sound
        this.playNotificationSound();
        
        // Show browser notification if not focused on chat
        if (document.hidden || this.currentChatUser !== messageData.senderId) {
            this.showBrowserNotification(messageData);
        }
        
        // Update conversations list
        this.loadConversations();
        
        // If chat window is open for this user, add message to chat
        if (this.currentChatWindow && this.currentChatUser === messageData.senderId) {
            this.displayMessageInChat(messageData);
        }
    }

    handleMessageRead(data) {
        console.log('📖 Message read notification:', data);
        // Update UI to show message as read
    }

    handleUserStatus(data) {
        console.log('👤 User status update:', data);
        this.updateUserOnlineStatus(data.userId, data.status === 'online');
    }

    async loadConversations() {
        try {
            console.log('🔄 Loading conversations...');
            const response = await fetch('/api/conversations');
            console.log('📡 Conversations response status:', response.status);

            const data = await response.json();
            console.log('📡 Conversations response data:', data);

            if (data.success) {
                this.conversations = data.data || [];
                this.renderConversations();
                console.log('✅ Conversations loaded:', this.conversations.length);
            } else {
                console.error('❌ Failed to load conversations:', data.error);
                this.renderConversationsError(data.error || 'Failed to load conversations');
            }
        } catch (error) {
            console.error('❌ Error loading conversations:', error);
            this.renderConversationsError('Network error loading conversations');
        }
    }

    async loadOnlineUsers() {
        try {
            console.log('🔄 Loading online users...');
            const response = await fetch('/api/online-users');
            console.log('📡 Online users response status:', response.status);

            const data = await response.json();
            console.log('📡 Online users response data:', data);

            if (data.success) {
                this.onlineUsers = data.data || [];
                this.renderOnlineUsers();
                console.log('✅ Online users loaded:', this.onlineUsers.length);
            } else {
                console.error('❌ Failed to load online users:', data.error);
                this.renderOnlineUsersError(data.error || 'Failed to load online users');
            }
        } catch (error) {
            console.error('❌ Error loading online users:', error);
            this.renderOnlineUsersError('Network error loading online users');
        }
    }

    renderConversations() {
        const container = document.getElementById('conversationsContainer');
        if (!container) return;

        if (this.conversations.length === 0) {
            container.innerHTML = `
                <div class="no-conversations">
                    <p>No conversations yet. Start chatting with online users!</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.conversations.map(conv => `
            <div class="conversation-item" onclick="window.messagesPage.openChat('${conv.otherUserId}', '${conv.otherUserNickname}')">
                <div class="conversation-avatar">
                    ${conv.otherUserAvatar ?
                        `<img src="${conv.otherUserAvatar}" alt="${conv.otherUserNickname}">` :
                        `<div class="avatar-placeholder">${conv.otherUserNickname.charAt(0).toUpperCase()}</div>`
                    }
                    ${conv.isOnline ? '<div class="online-indicator"></div>' : ''}
                </div>
                <div class="conversation-info">
                    <div class="conversation-header">
                        <span class="conversation-name">${conv.otherUserNickname}</span>
                        <span class="conversation-time" title="${this.formatFullTime(conv.lastMessageTime)}">${this.formatConversationTime(conv.lastMessageTime)}</span>
                    </div>
                    <div class="conversation-preview">
                        <span class="last-message">${conv.lastMessage || 'No messages yet'}</span>
                        ${conv.unreadCount > 0 ? `<span class="unread-badge">${conv.unreadCount}</span>` : ''}
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderOnlineUsers() {
        const container = document.getElementById('onlineUsersContainer');
        if (!container) return;

        const filteredUsers = this.onlineUsers.filter(user => user.userId !== this.currentUser.id);

        if (filteredUsers.length === 0) {
            container.innerHTML = '<div class="no-users">No other users online</div>';
            return;
        }

        container.innerHTML = filteredUsers.map(user => `
            <div class="online-user-item" onclick="window.messagesPage.openChat('${user.userId}', '${user.nickname}')">
                <div class="user-avatar">
                    ${user.avatarUrl ? 
                        `<img src="${user.avatarUrl}" alt="${user.nickname}">` :
                        `<div class="avatar-placeholder">${user.nickname.charAt(0).toUpperCase()}</div>`
                    }
                    <div class="online-indicator"></div>
                </div>
                <div class="user-info">
                    <span class="user-name">${user.nickname}</span>
                    <span class="user-status">Online</span>
                </div>
            </div>
        `).join('');
    }

    renderConversationsError(errorMessage) {
        const container = document.getElementById('conversationsContainer');
        if (!container) return;

        container.innerHTML = `
            <div class="error-message">
                <p>❌ ${errorMessage}</p>
                <button onclick="window.messagesPage.loadConversations()" class="retry-btn">Retry</button>
            </div>
        `;
    }

    renderOnlineUsersError(errorMessage) {
        const container = document.getElementById('onlineUsersContainer');
        if (!container) return;

        container.innerHTML = `
            <div class="error-message">
                <p>❌ ${errorMessage}</p>
                <button onclick="window.messagesPage.loadOnlineUsers()" class="retry-btn">Retry</button>
            </div>
        `;
    }

    async openChat(userId, nickname) {
        console.log(`💬 Opening chat with ${nickname} (${userId})`);
        
        this.currentChatUser = userId;
        
        // Close existing chat window
        if (this.currentChatWindow) {
            this.currentChatWindow.remove();
        }
        
        // Create new chat window
        this.currentChatWindow = this.createChatWindow(userId, nickname);
        document.body.appendChild(this.currentChatWindow);
        
        // Load messages (initial load - should show only last 10)
        console.log(`🔄 Opening chat with ${userId}, loading initial messages...`);
        await this.loadChatMessages(userId);
        
        // Focus on input
        const input = this.currentChatWindow.querySelector('.message-input');
        if (input) {
            setTimeout(() => input.focus(), 100);
        }
        
        // Mark messages as read
        await this.markMessagesAsRead(userId);
    }

    createChatWindow(userId, nickname) {
        const chatWindow = document.createElement('div');
        chatWindow.className = 'chat-window';
        chatWindow.innerHTML = `
            <div class="chat-header">
                <div class="chat-user-info">
                    <span class="chat-user-name">${nickname}</span>
                    <span class="chat-user-status" id="chatUserStatus-${userId}">Online</span>
                </div>
                <button class="chat-close-btn" onclick="window.messagesPage.closeChat()">×</button>
            </div>
            <div class="chat-messages" id="chatMessages-${userId}"></div>
            <div class="chat-input-container">
                <input type="text" class="message-input" placeholder="Type a message..." 
                       onkeypress="window.messagesPage.handleMessageKeyPress(event, '${userId}')">
                <button class="send-btn" onclick="window.messagesPage.sendMessage('${userId}')">Send</button>
            </div>
        `;
        
        return chatWindow;
    }

    async loadChatMessages(userId, isLoadMore = false) {
        try {
            // Initialize pagination for this user if not exists
            if (!this.chatPagination.has(userId)) {
                this.chatPagination.set(userId, { offset: 0, hasMore: true, loading: false });
            }

            const pagination = this.chatPagination.get(userId);

            // Prevent multiple simultaneous requests
            if (pagination.loading) {
                return;
            }

            // If no more messages available, return
            if (isLoadMore && !pagination.hasMore) {
                return;
            }

            pagination.loading = true;

            const limit = 10; // Load 10 messages at a time
            const offset = isLoadMore ? pagination.offset : 0;

            console.log(`📥 Loading messages for ${userId}: limit=${limit}, offset=${offset}, isLoadMore=${isLoadMore}`);

            const response = await fetch(`/api/messages?user=${userId}&limit=${limit}&offset=${offset}`);
            const data = await response.json();

            if (data.success) {
                const messages = data.data || [];

                // Update pagination
                pagination.offset = isLoadMore ? pagination.offset + messages.length : messages.length;
                pagination.hasMore = messages.length === limit; // If we got fewer than limit, no more messages
                pagination.loading = false;

                if (isLoadMore) {
                    this.prependMessages(messages, userId);
                } else {
                    this.displayMessages(messages, userId);
                }

                console.log(`✅ Loaded ${messages.length} messages for user ${userId} (total offset: ${pagination.offset})`);
            } else {
                console.error('❌ Failed to load messages:', data.error);
                pagination.loading = false;
            }
        } catch (error) {
            console.error('❌ Error loading messages:', error);
            if (this.chatPagination.has(userId)) {
                this.chatPagination.get(userId).loading = false;
            }
        }
    }

    displayMessages(messages, userId) {
        const container = document.getElementById(`chatMessages-${userId}`);
        if (!container) return;

        console.log(`📄 Displaying ${messages.length} messages for user ${userId}`);

        // API already returns messages in chronological order (oldest first)
        // No need to reverse - just display them as-is

        let html = '';
        let lastDate = null;
        let lastSender = null;

        messages.forEach((msg) => {
            const msgDate = new Date(msg.createdAt).toDateString();

            // Add date separator if date changed
            if (lastDate && lastDate !== msgDate) {
                html += this.createDateSeparator(msgDate);
            } else if (!lastDate) {
                html += this.createDateSeparator(msgDate);
            }

            // Check if this is a consecutive message from same sender
            const isSameSender = lastSender === msg.senderId;
            const messageHTML = this.createMessageHTML(msg, isSameSender);

            // Add sender ID as data attribute
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = messageHTML;
            const messageElement = tempDiv.querySelector('.message');
            if (messageElement) {
                messageElement.dataset.senderId = msg.senderId;
                messageElement.dataset.messageId = msg.id;
                messageElement.dataset.createdAt = msg.createdAt;
            }
            html += tempDiv.innerHTML;

            lastDate = msgDate;
            lastSender = msg.senderId;
        });

        container.innerHTML = html;

        // Add scroll event listener for pagination
        this.addScrollListener(container, userId);

        // Scroll to bottom to show most recent messages
        container.scrollTop = container.scrollHeight;
    }

    prependMessages(messages, userId) {
        const container = document.getElementById(`chatMessages-${userId}`);
        if (!container) return;

        if (messages.length === 0) return;

        // Store current scroll position
        const scrollHeight = container.scrollHeight;
        const scrollTop = container.scrollTop;

        // API already returns messages in chronological order (oldest first)
        // No need to reverse - just display them as-is

        // Get the first existing message to check for date separator needs
        const firstExistingMessage = container.querySelector('.message');
        let needsDateSeparatorAfter = false;

        if (firstExistingMessage && messages.length > 0) {
            const firstExistingDate = new Date(firstExistingMessage.dataset.createdAt).toDateString();
            const lastNewMessageDate = new Date(messages[messages.length - 1].createdAt).toDateString();
            needsDateSeparatorAfter = firstExistingDate !== lastNewMessageDate;
        }

        let html = '';
        let lastDate = null;
        let lastSender = null;

        messages.forEach((msg, index) => {
            const msgDate = new Date(msg.createdAt).toDateString();

            // Add date separator if date changed
            if (lastDate && lastDate !== msgDate) {
                html += this.createDateSeparator(msgDate);
            } else if (!lastDate) {
                html += this.createDateSeparator(msgDate);
            }

            // Check if this is a consecutive message from same sender
            const isSameSender = lastSender === msg.senderId;
            const messageHTML = this.createMessageHTML(msg, isSameSender);

            // Add message data attributes
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = messageHTML;
            const messageElement = tempDiv.querySelector('.message');
            if (messageElement) {
                messageElement.dataset.senderId = msg.senderId;
                messageElement.dataset.messageId = msg.id;
                messageElement.dataset.createdAt = msg.createdAt;
            }
            html += tempDiv.innerHTML;

            lastDate = msgDate;
            lastSender = msg.senderId;
        });

        // Add date separator between old and new messages if needed
        if (needsDateSeparatorAfter && firstExistingMessage) {
            const firstExistingDate = new Date(firstExistingMessage.dataset.createdAt).toDateString();
            html += this.createDateSeparator(firstExistingDate);
        }

        // Prepend the older messages
        container.insertAdjacentHTML('afterbegin', html);

        // Restore scroll position (maintain relative position)
        const newScrollHeight = container.scrollHeight;
        container.scrollTop = scrollTop + (newScrollHeight - scrollHeight);
    }

    addScrollListener(container, userId) {
        // Remove existing listener if any
        if (container.scrollHandler) {
            container.removeEventListener('scroll', container.scrollHandler);
        }

        // Create throttled scroll handler
        const scrollHandler = this.throttle((event) => {
            const container = event.target;
            const scrollTop = container.scrollTop;
            const scrollThreshold = 100; // Load more when within 100px of top

            // Check if scrolled near the top
            if (scrollTop <= scrollThreshold) {
                console.log(`📜 Scroll threshold reached for user ${userId}, loading more messages...`);
                this.loadChatMessages(userId, true); // Load more messages
            }
        }, 300); // Throttle to 300ms

        // Store reference for cleanup
        container.scrollHandler = scrollHandler;

        // Add event listener
        container.addEventListener('scroll', scrollHandler);
    }

    createMessageHTML(message, isSameSender = false) {
        const isOwnMessage = message.senderId === this.currentUser.id;
        const time = this.formatMessageTime(message.createdAt);
        const fullTime = this.formatFullTime(message.createdAt);

        return `
            <div class="message ${isOwnMessage ? 'own-message' : 'other-message'} ${isSameSender ? 'same-sender' : ''}">
                <div class="message-header">
                    ${!isOwnMessage && !isSameSender ? `<span class="message-sender">${message.senderNickname}</span>` : ''}
                    <span class="message-timestamp" title="${fullTime}">${time}</span>
                </div>
                <div class="message-content">${this.escapeHtml(message.content)}</div>
            </div>
        `;
    }

    createDateSeparator(dateString) {
        const date = new Date(dateString);
        const today = new Date().toDateString();
        const yesterday = new Date(Date.now() - 86400000).toDateString();

        let displayDate;
        if (dateString === today) {
            displayDate = 'Today';
        } else if (dateString === yesterday) {
            displayDate = 'Yesterday';
        } else {
            displayDate = date.toLocaleDateString([], {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            });
        }

        return `
            <div class="date-separator">
                <span>${displayDate}</span>
            </div>
        `;
    }

    displayMessageInChat(message) {
        const container = document.getElementById(`chatMessages-${this.currentChatUser}`);
        if (!container) return;

        // Update pagination offset to account for new message
        if (this.chatPagination.has(this.currentChatUser)) {
            const pagination = this.chatPagination.get(this.currentChatUser);
            pagination.offset += 1;
        }

        // Check if we need a date separator
        const lastMessage = container.querySelector('.message:last-child');
        let needsDateSeparator = false;

        if (lastMessage) {
            const lastTimestamp = lastMessage.querySelector('.message-timestamp')?.title;
            if (lastTimestamp) {
                const lastDate = new Date(lastTimestamp).toDateString();
                const currentDate = new Date(message.createdAt).toDateString();
                needsDateSeparator = lastDate !== currentDate;
            }
        }

        // Check if this is from the same sender as the last message
        const lastSenderId = lastMessage?.dataset?.senderId;
        const isSameSender = lastSenderId === message.senderId && !needsDateSeparator;

        let html = '';
        if (needsDateSeparator) {
            html += this.createDateSeparator(new Date(message.createdAt).toDateString());
        }

        html += this.createMessageHTML(message, isSameSender);

        // Add message data attributes
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = html;
        const messageElement = tempDiv.querySelector('.message');
        if (messageElement) {
            messageElement.dataset.senderId = message.senderId;
            messageElement.dataset.messageId = message.id;
            messageElement.dataset.createdAt = message.createdAt;
        }

        container.insertAdjacentHTML('beforeend', tempDiv.innerHTML);

        // Scroll to bottom
        container.scrollTop = container.scrollHeight;
    }

    handleMessageKeyPress(event, userId) {
        if (event.key === 'Enter') {
            this.sendMessage(userId);
        }
    }

    async sendMessage(userId) {
        const input = this.currentChatWindow.querySelector('.message-input');
        const content = input.value.trim();
        
        if (!content) return;
        
        try {
            const response = await fetch('/api/messages/send', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    receiverId: userId,
                    content: content
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                // Clear input
                input.value = '';
                
                // Display message immediately
                this.displayMessageInChat(data.data);
                
                // Update conversations
                this.loadConversations();
                
                console.log('✅ Message sent successfully');
            } else {
                console.error('❌ Failed to send message:', data.error);
                alert('Failed to send message: ' + data.error);
            }
        } catch (error) {
            console.error('❌ Error sending message:', error);
            alert('Error sending message. Please try again.');
        }
    }

    async markMessagesAsRead(senderId) {
        try {
            await fetch('/api/messages/read', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    senderId: senderId
                })
            });
        } catch (error) {
            console.error('❌ Error marking messages as read:', error);
        }
    }

    closeChat() {
        if (this.currentChatWindow) {
            // Clean up scroll listener
            const container = this.currentChatWindow.querySelector('[id^="chatMessages-"]');
            if (container && container.scrollHandler) {
                container.removeEventListener('scroll', container.scrollHandler);
                container.scrollHandler = null;
            }

            // Reset pagination for this user
            if (this.currentChatUser) {
                this.chatPagination.delete(this.currentChatUser);
            }

            this.currentChatWindow.remove();
            this.currentChatWindow = null;
            this.currentChatUser = null;
        }
    }

    setupEventListeners() {
        // Refresh button
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadConversations();
                this.loadOnlineUsers();
            });
        }

        // Close chat on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.currentChatWindow) {
                this.closeChat();
            }
        });
    }

    initNotificationSound() {
        // Create a simple notification sound
        try {
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        } catch (e) {
            console.log('Web Audio API not supported');
        }
    }

    playNotificationSound() {
        if (!this.audioContext) return;
        
        try {
            const oscillator = this.audioContext.createOscillator();
            const gainNode = this.audioContext.createGain();
            
            oscillator.connect(gainNode);
            gainNode.connect(this.audioContext.destination);
            
            oscillator.frequency.setValueAtTime(800, this.audioContext.currentTime);
            gainNode.gain.setValueAtTime(0.1, this.audioContext.currentTime);
            gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioContext.currentTime + 0.1);
            
            oscillator.start(this.audioContext.currentTime);
            oscillator.stop(this.audioContext.currentTime + 0.1);
        } catch (e) {
            console.log('Error playing notification sound:', e);
        }
    }

    showBrowserNotification(message) {
        if (Notification.permission === 'granted') {
            const notification = new Notification(`New message from ${message.senderNickname}`, {
                body: message.content.length > 50 ? message.content.substring(0, 50) + '...' : message.content,
                icon: '/static/images/message-icon.png',
                tag: `message-${message.senderId}`
            });
            
            notification.onclick = () => {
                window.focus();
                this.openChat(message.senderId, message.senderNickname);
                notification.close();
            };
            
            setTimeout(() => notification.close(), 5000);
        } else if (Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }

    updateUserOnlineStatus(userId, isOnline) {
        // Update online users list
        if (isOnline) {
            // User came online - refresh online users
            this.loadOnlineUsers();
        } else {
            // User went offline - remove from online users
            this.onlineUsers = this.onlineUsers.filter(user => user.userId !== userId);
            this.renderOnlineUsers();
        }
        
        // Update chat status if chat is open
        const statusElement = document.getElementById(`chatUserStatus-${userId}`);
        if (statusElement) {
            statusElement.textContent = isOnline ? 'Online' : 'Offline';
            statusElement.className = `chat-user-status ${isOnline ? 'online' : 'offline'}`;
        }
        
        // Update conversations list
        this.loadConversations();
    }

    formatTime(timestamp) {
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

    formatMessageTime(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        // For messages, show more precise time
        if (diffMins < 1) return 'now';
        if (diffMins < 60) return `${diffMins}m`;
        if (diffHours < 24) {
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        }
        if (diffDays === 1) return `Yesterday ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
        if (diffDays < 7) {
            const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
            return `${dayNames[date.getDay()]} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
        }

        return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }

    formatFullTime(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString([], {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }

    formatConversationTime(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        // For conversation list, show concise but informative time
        if (diffMins < 1) return 'now';
        if (diffMins < 60) return `${diffMins}m`;
        if (diffHours < 24) return `${diffHours}h`;
        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) {
            const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
            return dayNames[date.getDay()];
        }

        return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }

    showTab(tabName, targetElement) {
        // Update tab buttons
        document.querySelectorAll('.sidebar-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        if (targetElement) {
            targetElement.classList.add('active');
        }

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.style.display = 'none';
        });

        if (tabName === 'conversations') {
            document.getElementById('conversationsContainer').style.display = 'block';
        } else if (tabName === 'online') {
            document.getElementById('onlineUsersContainer').style.display = 'block';
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize the messages page
let messagesPage;

function initMessagesPage() {
    messagesPage = new MessagesPage();
}

// Export for use in router
window.initMessagesPage = initMessagesPage;

// Export MessagesPage class for router
window.MessagesPage = {
    render: async function() {
        const container = document.getElementById('main-content');
        if (!container) return;

        container.innerHTML = `
            <div class="messaging-container">
                <div class="messaging-sidebar">
                    <div class="sidebar-header">
                        <h2>Messages</h2>
                    </div>
                    <div class="sidebar-tabs">
                        <button class="sidebar-tab active" onclick="window.messagesPage.showTab('conversations', this)">Conversations</button>
                        <button class="sidebar-tab" onclick="window.messagesPage.showTab('online', this)">Online Users</button>
                    </div>
                    <div class="sidebar-content">
                        <div id="conversationsContainer" class="tab-content active">
                            <div class="loading">Loading conversations...</div>
                        </div>
                        <div id="onlineUsersContainer" class="tab-content" style="display: none;">
                            <div class="loading">Loading online users...</div>
                        </div>
                    </div>
                </div>
                <div class="messaging-main">
                    <div class="messaging-welcome">
                        <h3>Welcome to Messages</h3>
                        <p>Select a conversation or start chatting with online users</p>
                    </div>
                </div>
            </div>
        `;

        // Initialize the messages page
        if (!window.messagesPage) {
            window.messagesPage = new MessagesPage();
        }

        // Always call init to handle re-entry
        await window.messagesPage.init();
    }
};
