// API client for the Forum application

window.api = {
    baseURL: '/api',

    /**
     * Make HTTP request
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            credentials: 'include', // Include cookies for session management
            ...options
        };

        try {
            console.log('📡 Making request to:', url);
            console.log('📡 Request config:', config);

            const response = await fetch(url, config);
            console.log('📡 Response status:', response.status);
            console.log('📡 Response ok:', response.ok);

            const data = await response.json();
            console.log('📡 Response data:', data);

            if (!response.ok) {
                const errorMessage = data.error || data.message || `HTTP error! status: ${response.status}`;

                // Show custom error page for certain status codes
                if (response.status === 400) {
                    window.showErrorPage(400, errorMessage);
                } else if (response.status === 500) {
                    window.showErrorPage(500, errorMessage);
                }

                throw new Error(errorMessage);
            }

            return data;
        } catch (error) {
            console.error('📡 API request failed:', error);
            throw error;
        }
    },

    /**
     * GET request
     */
    async get(endpoint, params = {}) {
        let url = endpoint;
        if (Object.keys(params).length > 0) {
            const searchParams = new URLSearchParams();
            Object.keys(params).forEach(key => {
                if (params[key] !== null && params[key] !== undefined) {
                    searchParams.append(key, params[key]);
                }
            });
            url += '?' + searchParams.toString();
        }

        return this.request(url, {
            method: 'GET'
        });
    },

    /**
     * POST request
     */
    async post(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    /**
     * PUT request
     */
    async put(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    /**
     * DELETE request
     */
    async delete(endpoint) {
        return this.request(endpoint, {
            method: 'DELETE'
        });
    },

    // Authentication endpoints
    async register(userData) {
        return this.post('/register', userData);
    },

    async login(identifier, password) {
        return this.post('/login', { identifier, password });
    },

    async logout() {
        return this.post('/logout');
    },

    async getCurrentUser() {
        return this.get('/user');
    },

    // Posts endpoints
    async getPosts(params = {}) {
        return this.get('/posts', params);
    },

    async getPost(postId) {
        return this.get(`/posts/${postId}`);
    },

    async createPost(postData) {
        return this.post('/posts', postData);
    },

    async likePost(postId, isLike) {
        return this.post('/like', { postId, isLike });
    },

    // Comments endpoints
    async getComments(postId) {
        console.log('📡 API: Getting comments for post:', postId);
        const url = `/comments/${postId}`;
        console.log('📡 API: Request URL:', url);
        return this.get(url);
    },

    async createComment(commentData) {
        return this.post('/comment', commentData);
    },

    async likeComment(commentId, isLike) {
        return this.post('/like', { commentId, isLike });
    },

    async updateComment(commentId, data) {
        return this.put(`/comment/${commentId}`, data);
    },

    async deleteComment(commentId) {
        return this.delete(`/comment/${commentId}`);
    },



    // Users endpoints
    async getUsers(params = {}) {
        return this.get('/users', params);
    },

    async getOnlineUsers() {
        return this.get('/online-users');
    },

    // Messaging endpoints
    async getConversations() {
        return this.get('/conversations');
    },

    async getMessages(conversationId, page = 1, limit = 50) {
        return this.get(`/messages/${conversationId}?page=${page}&limit=${limit}`);
    },

    async sendMessage(recipientId, content, messageType = 'text') {
        return this.post('/messages', { recipientId, content, messageType });
    },

    async markAsRead(conversationId) {
        return this.put(`/messages/${conversationId}/read`);
    },

    async searchUsers(query) {
        return this.get(`/users/search?q=${encodeURIComponent(query)}`);
    },

    // Profile endpoints
    async getProfile() {
        return this.get('/profile');
    },

    async updateProfile(profileData) {
        return this.put('/profile', profileData);
    },

    async uploadAvatar(formData) {
        return this.request('/upload/avatar', {
            method: 'POST',
            body: formData,
            headers: {} // Remove Content-Type header to let browser set it for FormData
        });
    },

    async updateAvatar(avatarData) {
        return this.put('/profile/avatar', avatarData);
    },



    // Categories endpoints
    async getCategories() {
        return this.get('/categories');
    }
};

// Global error handler for API errors
window.handleAPIError = function(error, defaultMessage = 'An error occurred') {
    console.error('API Error:', error);
    
    let message = defaultMessage;
    if (error.message) {
        message = error.message;
    }
    
    // Show notification if available
    if (window.forumApp && window.forumApp.notificationComponent) {
        window.forumApp.notificationComponent.error(message);
    } else {
        // Fallback to alert
        alert(message);
    }
};

// Request interceptor for authentication
const originalRequest = window.api.request;
window.api.request = async function(endpoint, options = {}) {
    try {
        return await originalRequest.call(this, endpoint, options);
    } catch (error) {
        console.log('🔍 API request error:', error.message, 'for endpoint:', endpoint);

        // Handle authentication errors
        if (error.message.includes('Authentication required') ||
            error.message.includes('Not authenticated') ||
            error.message.includes('No active session')) {

            console.log('🔒 Authentication error detected, clearing local state');

            // Clear any stored user data
            if (window.auth) {
                window.auth.currentUser = null;
                window.auth.isAuthenticated = false;
            }

            if (window.forumApp) {
                window.forumApp.currentUser = null;
                window.forumApp.isAuthenticated = false;
                window.forumApp.updateAuthUI();
            }
            
            // Redirect to login if not already there
            if (window.location.pathname !== '/login' && 
                window.location.pathname !== '/register') {
                if (window.forumApp && window.forumApp.router) {
                    window.forumApp.router.navigate('/login');
                } else {
                    window.location.href = '/login';
                }
            }
        }
        
        throw error;
    }
};
