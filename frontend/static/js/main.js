// Main application entry point

window.forumApp = {
    // Core components
    router: null,
    currentUser: null,
    isAuthenticated: false,
    currentPage: null,

    // UI components
    headerComponent: null,
    sidebarComponent: null,
    chatComponent: null,
    notificationComponent: null,

    // Page components
    pages: {},

    /**
     * Initialize the application
     */
    async init() {
        console.log('🚀 Initializing Forum Application...');
        
        try {
            // Show loading screen
            this.showLoading();

            // Initialize authentication
            await this.initAuth();

            // Initialize components
            this.initComponents();

            // Initialize WebSocket for real-time features (online users, notifications)
            if (this.isAuthenticated) {
                this.initWebSocket();
            }

            // Update UI based on auth state
            this.updateAuthUI();

            // Hide loading screen and show app
            this.hideLoading();

            console.log('✅ Forum Application initialized successfully');
            
        } catch (error) {
            console.error('❌ Failed to initialize application:', error);
            this.showError('Failed to initialize application. Please refresh the page.');
        }
    },

    /**
     * Initialize authentication
     */
    async initAuth() {
        const isAuthenticated = await window.auth.init();
        this.currentUser = window.auth.getCurrentUser();
        this.isAuthenticated = isAuthenticated;
        
        if (isAuthenticated) {
            console.log('✅ User authenticated:', this.currentUser.nickname);
        } else {
            console.log('ℹ️ User not authenticated');
        }
    },

    /**
     * Initialize UI components
     */
    initComponents() {
        // Initialize notification component first (other components may use it)
        if (window.NotificationComponent) {
            this.notificationComponent = window.NotificationComponent;
            this.notificationComponent.init();
        }

        // Initialize header component
        if (window.HeaderComponent) {
            this.headerComponent = window.HeaderComponent;
            this.headerComponent.init();
        }

        // Initialize sidebar component
        if (window.SidebarComponent) {
            console.log('🔧 Initializing sidebar component...');
            this.sidebarComponent = window.SidebarComponent;
            this.sidebarComponent.init();
        } else {
            console.error('❌ SidebarComponent not found!');
        }

        // Initialize chat component
        if (window.ChatComponent) {
            this.chatComponent = window.ChatComponent;
            this.chatComponent.init();
        }

        // Initialize page components
        this.initPages();

        // Setup error handling
        this.setupErrorHandling();
    },

    setupErrorHandling() {
        // Handle fetch errors globally
        const originalFetch = window.fetch;
        window.fetch = async function(...args) {
            try {
                const response = await originalFetch.apply(this, args);

                // Check for HTTP error status codes
                if (!response.ok) {
                    if (response.status === 404) {
                        window.showErrorPage(404, 'Page not found');
                        return response;
                    } else if (response.status === 500) {
                        window.showErrorPage(500, 'Internal server error');
                        return response;
                    } else if (response.status === 400) {
                        window.showErrorPage(400, 'Bad request');
                        return response;
                    }
                }

                return response;
            } catch (error) {
                console.error('Network error:', error);
                window.showErrorPage(500, 'Network error occurred');
                throw error;
            }
        };
    },

    /**
     * Initialize page components
     */
    initPages() {
        this.pages = {
            home: window.HomePage ? window.HomePage.render.bind(window.HomePage) : this.defaultPageHandler('Home'),
            login: window.LoginPage ? window.LoginPage.render.bind(window.LoginPage) : this.defaultPageHandler('Login'),
            register: window.RegisterPage ? window.RegisterPage.render.bind(window.RegisterPage) : this.defaultPageHandler('Register'),
            posts: window.PostsPage ? window.PostsPage.render.bind(window.PostsPage) : this.defaultPageHandler('Posts'),
            post: window.PostPage ? window.PostPage.render.bind(window.PostPage) : this.defaultPageHandler('Post'),
            createPost: window.CreatePostPage ? window.CreatePostPage.render.bind(window.CreatePostPage) : this.defaultPageHandler('Create Post'),
            messages: window.MessagesPage ? window.MessagesPage.render.bind(window.MessagesPage) : this.defaultPageHandler('Messages'),
            profile: window.ProfilePage ? window.ProfilePage.render.bind(window.ProfilePage) : this.defaultPageHandler('Profile'),

        };

        // Initialize router after pages are set up
        this.initRouter();
    },

    /**
     * Default page handler for missing pages
     */
    defaultPageHandler(pageName) {
        return () => {
            const mainContent = document.getElementById('main-content');
            mainContent.innerHTML = `
                <div class="text-center">
                    <h1>${pageName}</h1>
                    <p>This page is under development.</p>
                </div>
            `;
        };
    },

    /**
     * Initialize router
     */
    initRouter() {
        this.router = new window.Router();

        // Register routes - all require authentication except login/register
        this.router.addRoute('/', this.pages.home, {
            title: 'Forum - Home',
            requiresAuth: true
        });

        this.router.addRoute('/login', this.pages.login, {
            title: 'Forum - Login',
            redirectIfAuth: true
        });

        this.router.addRoute('/register', this.pages.register, {
            title: 'Forum - Register',
            redirectIfAuth: true
        });

        this.router.addRoute('/posts', this.pages.posts, {
            title: 'Forum - Posts',
            requiresAuth: true
        });

        this.router.addRoute('/post/:id', this.pages.post, {
            title: 'Forum - Post',
            requiresAuth: true
        });

        this.router.addRoute('/create-post', this.pages.createPost, {
            title: 'Forum - Create Post',
            requiresAuth: true
        });

        this.router.addRoute('/messages', this.pages.messages, {
            title: 'Forum - Messages',
            requiresAuth: true
        });

        this.router.addRoute('/profile', this.pages.profile, {
            title: 'Forum - Profile',
            requiresAuth: true
        });



        // Error test routes
        this.router.addRoute('/test-404', () => {
            window.showErrorPage(404, 'This is a test 404 error page');
        }, {
            title: 'Forum - Test 404',
            requiresAuth: true
        });

        this.router.addRoute('/test-500', () => {
            window.showErrorPage(500, 'This is a test 500 error page');
        }, {
            title: 'Forum - Test 500',
            requiresAuth: true
        });

        this.router.addRoute('/test-400', () => {
            window.showErrorPage(400, 'This is a test 400 error page');
        }, {
            title: 'Forum - Test 400',
            requiresAuth: true
        });

        // 404 route
        this.router.addRoute('/404', () => {
            window.showErrorPage(404);
        }, {
            title: 'Forum - Page Not Found',
            requiresAuth: true
        });

        // Initialize router
        this.router.init();
    },

    /**
     * Initialize WebSocket connection for real-time features
     */
    initWebSocket() {
        if (!this.isAuthenticated) {
            console.log('🔌 Skipping WebSocket initialization - user not authenticated');
            return;
        }

        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws`;

        console.log('🔌 Initializing main WebSocket connection:', wsUrl);

        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
            console.log('✅ Main WebSocket connected');
            this.isWebSocketConnected = true;
        };

        this.websocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('❌ Error parsing WebSocket message:', error);
            }
        };

        this.websocket.onclose = () => {
            console.log('🔌 Main WebSocket disconnected');
            this.isWebSocketConnected = false;

            // Attempt to reconnect if user is still authenticated
            if (this.isAuthenticated) {
                console.log('🔄 Attempting to reconnect WebSocket in 3 seconds...');
                setTimeout(() => {
                    if (this.isAuthenticated) {
                        this.initWebSocket();
                    }
                }, 3000);
            }
        };

        this.websocket.onerror = (error) => {
            console.error('❌ Main WebSocket error:', error);
        };
    },

    /**
     * Handle incoming WebSocket messages
     */
    handleWebSocketMessage(message) {
        console.log('📨 Received WebSocket message:', message);

        switch (message.type) {
            case 'user_status':
                this.handleUserStatusUpdate(message.data);
                // Also forward to messages page for online user updates
                if (window.messagesPage && window.messagesPage.handleWebSocketMessage) {
                    window.messagesPage.handleWebSocketMessage(message);
                }
                break;
            case 'new_post':
                this.handleNewPost(message.data);
                break;
            case 'notification':
                this.handleNotification(message.data);
                break;
            case 'new_message':
            case 'message_read':
                // Forward messaging-related messages to the messages page if it exists
                console.log('📨 Main App: Forwarding message to messages page:', message.type);
                if (window.messagesPage && window.messagesPage.handleWebSocketMessage) {
                    console.log('📨 Main App: Messages page exists, forwarding...');
                    window.messagesPage.handleWebSocketMessage(message);
                } else {
                    console.log('📨 Main App: Messages page not available for forwarding');
                }
                break;
            default:
                console.log('🔍 Unknown WebSocket message type:', message.type);
        }
    },

    /**
     * Handle user status updates (online/offline)
     */
    handleUserStatusUpdate(data) {
        console.log('👥 User status update:', data);

        // Notify sidebar component about user status change
        if (window.SidebarComponent) {
            window.SidebarComponent.updateUserStatus(data);
        }

        // Dispatch custom event for other components
        window.dispatchEvent(new CustomEvent('user_status', { detail: data }));
    },

    /**
     * Handle new post notifications
     */
    handleNewPost(data) {
        console.log('📝 New post notification:', data);

        // Show notification if not on posts page
        if (this.router.currentRoute !== 'posts' && this.notificationComponent) {
            this.notificationComponent.info(`New post: ${data.title}`);
        }
    },

    /**
     * Handle general notifications
     */
    handleNotification(data) {
        console.log('🔔 Notification:', data);

        if (this.notificationComponent) {
            this.notificationComponent.info(data.message);
        }
    },

    /**
     * Update UI based on authentication state
     */
    updateAuthUI() {
        const header = document.getElementById('header');
        const mainContent = document.getElementById('main-content');
        const sidebar = document.getElementById('sidebar');
        const userMenu = document.getElementById('user-menu');
        const authButtons = document.getElementById('auth-buttons');
        const userNickname = document.getElementById('user-nickname');
        const userAvatar = document.getElementById('user-avatar');
        const navigation = document.querySelector('.nav-links');

        if (this.isAuthenticated && this.currentUser) {
            // Show authenticated UI - full app interface
            if (header) header.style.display = 'block';
            if (mainContent) mainContent.style.display = 'block';
            if (navigation) navigation.style.display = 'flex';
            if (userMenu) userMenu.style.display = 'block';
            if (authButtons) authButtons.style.display = 'none';
            if (sidebar) {
                sidebar.style.display = 'block';
                console.log('👥 Sidebar made visible');
            } else {
                console.error('❌ Sidebar element not found!');
            }

            // Update user info
            if (userNickname) userNickname.textContent = this.currentUser.nickname;
            if (userAvatar) {
                userAvatar.src = this.currentUser.avatarUrl || '/static/images/default-avatar.svg';
            }

            console.log('🔓 Showing authenticated UI');

        } else {
            // Show unauthenticated UI - hide everything except main content
            if (header) header.style.display = 'none';
            if (navigation) navigation.style.display = 'none';
            if (mainContent) mainContent.style.display = 'block';
            if (userMenu) userMenu.style.display = 'none';
            if (authButtons) authButtons.style.display = 'none';
            if (sidebar) sidebar.style.display = 'none';

            console.log('🔒 Showing unauthenticated UI - header and navbar hidden');
        }

        // Note: Logout handler is managed by HeaderComponent to avoid duplicate handlers
    },

    /**
     * Handle user logout
     */
    async logout() {
        console.log('🔒 Main: logout() called');
        try {
            await window.auth.logout();
            console.log('🔒 Main: auth.logout() completed');

            this.currentUser = null;
            this.isAuthenticated = false;

            // Close WebSocket connection
            if (this.websocket) {
                this.websocket.close();
                this.websocket = null;
                this.isWebSocketConnected = false;
            }

            // Update UI
            this.updateAuthUI();

            // Redirect to login page
            if (this.router) {
                this.router.navigate('/login');
            }

            console.log('🔒 Main: User logged out and redirected to login');

        } catch (error) {
            console.error('🔒 Main: Logout error:', error);
            if (this.notificationComponent) {
                this.notificationComponent.error('Failed to logout');
            }
        }
    },

    /**
     * Handle successful login/registration
     */
    onAuthSuccess(user) {
        this.currentUser = user;
        this.isAuthenticated = true;

        // Initialize WebSocket for real-time features
        this.initWebSocket();

        // Update UI
        this.updateAuthUI();

        // Restart sidebar to immediately fetch online users for real-time experience
        if (this.sidebarComponent) {
            console.log('👥 Restarting sidebar component after login');
            this.sidebarComponent.restart();
        }

        // Navigate to home page
        if (this.router) {
            this.router.navigate('/');
        }
    },

    /**
     * Show loading screen
     */
    showLoading() {
        const loadingScreen = document.getElementById('loading-screen');
        if (loadingScreen) {
            loadingScreen.style.display = 'flex';
        }
    },

    /**
     * Hide loading screen
     */
    hideLoading() {
        const loadingScreen = document.getElementById('loading-screen');
        if (loadingScreen) {
            loadingScreen.style.display = 'none';
        }
    },

    /**
     * Show error message
     */
    showError(message) {
        this.hideLoading();
        
        if (this.notificationComponent) {
            this.notificationComponent.error(message);
        } else {
            alert(message);
        }
    },

    /**
     * Set current page
     */
    setCurrentPage(pageName) {
        this.currentPage = pageName;

        // Handle sidebar visibility based on page
        this.updateSidebarVisibility(pageName);
    },

    /**
     * Update sidebar visibility based on current page
     */
    updateSidebarVisibility(pageName) {
        const sidebar = document.getElementById('sidebar');
        if (!sidebar) return;

        if (pageName === 'messages') {
            // Hide sidebar on messages page to give more space for chat
            sidebar.style.display = 'none';
            console.log('👥 Hiding right sidebar for messages page');
        } else if (this.isAuthenticated) {
            // Show sidebar on other pages if user is authenticated
            sidebar.style.display = 'block';
            console.log('👥 Showing right sidebar for page:', pageName);
        }
    }
};

// Simple Error Page Component
window.ErrorPage = {
    render(errorCode = 404, errorMessage = null) {
        window.forumApp.setCurrentPage('error');

        const errorInfo = this.getErrorInfo(errorCode);
        const customMessage = errorMessage || errorInfo.message;

        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="error-page">
                <div class="error-container">
                    <div class="error-icon">${errorInfo.icon}</div>
                    <h1 class="error-code">${errorCode}</h1>
                    <h2 class="error-title">${errorInfo.title}</h2>
                    <p class="error-message">${customMessage}</p>
                    <div class="error-actions">
                        ${this.getErrorActions(errorCode)}
                    </div>
                </div>
            </div>
        `;

        this.bindEvents();
    },

    getErrorInfo(code) {
        const errorTypes = {
            400: {
                title: 'Bad Request',
                message: 'The request could not be understood by the server.',
                icon: '⚠️'
            },
            404: {
                title: 'Page Not Found',
                message: 'The page you\'re looking for doesn\'t exist or has been moved.',
                icon: '🔍'
            },
            500: {
                title: 'Internal Server Error',
                message: 'Something went wrong on our end. We\'re working to fix it.',
                icon: '🔧'
            }
        };

        return errorTypes[code] || {
            title: 'Unknown Error',
            message: 'An unexpected error occurred.',
            icon: '❓'
        };
    },

    getErrorActions(code) {
        const isAuthenticated = window.forumApp && window.forumApp.isAuthenticated;

        switch (code) {
            case 404:
                return `
                    <a href="/" data-route="/" class="btn btn-primary">
                        🏠 Go Home
                    </a>
                    ${isAuthenticated ? `
                        <a href="/posts" data-route="/posts" class="btn btn-secondary">
                            📝 Browse Posts
                        </a>
                    ` : ''}
                    <button onclick="history.back()" class="btn btn-secondary">
                        ← Go Back
                    </button>
                `;
            default:
                return `
                    <button onclick="location.reload()" class="btn btn-primary">
                        🔄 Refresh Page
                    </button>
                    <a href="/" data-route="/" class="btn btn-secondary">
                        🏠 Go Home
                    </a>
                `;
        }
    },

    bindEvents() {
        const routeLinks = document.querySelectorAll('[data-route]');
        routeLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const route = link.getAttribute('data-route');
                if (window.forumApp && window.forumApp.router) {
                    window.forumApp.router.navigate(route);
                }
            });
        });
    }
};

// Simple global error handler
window.showErrorPage = function(errorCode, errorMessage = null) {
    console.log(`🚨 Showing error page: ${errorCode}`, errorMessage);
    console.log('🔧 ErrorPage available:', !!window.ErrorPage);

    if (window.ErrorPage) {
        console.log('✅ Rendering custom error page');
        window.ErrorPage.render(errorCode, errorMessage);
    } else {
        console.error('❌ ErrorPage not available');
    }
};

// Initialize application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.forumApp.init();
});

// Page visibility and online/offline handling removed - using dedicated messaging WebSocket only

window.addEventListener('offline', () => {
    console.log('📴 Gone offline');
    if (window.forumApp.notificationComponent) {
        window.forumApp.notificationComponent.warning('You are offline');
    }
});
