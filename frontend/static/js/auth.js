// Authentication module for the Forum application

window.auth = {
    currentUser: null,
    isAuthenticated: false,

    /**
     * Initialize authentication
     */
    async init() {
        try {
            const response = await window.api.getCurrentUser();
            if (response.success && response.data) {
                this.currentUser = response.data;
                this.isAuthenticated = true;
                console.log('✅ User authenticated:', this.currentUser.nickname);
            }
        } catch (error) {
            console.log('ℹ️ User not authenticated');
            this.currentUser = null;
            this.isAuthenticated = false;
        }
        
        return this.isAuthenticated;
    },

    /**
     * Register a new user
     */
    async register(userData) {
        try {
            // Validate input
            const validation = this.validateRegistrationData(userData);
            if (!validation.isValid) {
                throw new Error(validation.message);
            }

            const response = await window.api.register(userData);
            
            if (response.success && response.data) {
                this.currentUser = response.data;
                this.isAuthenticated = true;
                
                // Show success notification
                if (window.forumApp && window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success('Registration successful! Welcome to Forum.');
                }
                
                return { success: true, user: this.currentUser };
            } else {
                throw new Error(response.error || 'Registration failed');
            }
        } catch (error) {
            console.error('Registration error:', error);
            throw error;
        }
    },

    /**
     * Login user
     */
    async login(identifier, password) {
        try {
            // Validate input
            if (!identifier || !password) {
                throw new Error('Email/nickname and password are required');
            }

            const response = await window.api.login(identifier, password);
            
            if (response.success && response.data) {
                this.currentUser = response.data;
                this.isAuthenticated = true;
                
                // Show success notification
                if (window.forumApp && window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success(`Welcome back, ${this.currentUser.nickname}!`);
                }
                
                return { success: true, user: this.currentUser };
            } else {
                throw new Error(response.error || 'Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    },

    /**
     * Logout user
     */
    async logout() {
        console.log('🔒 Starting logout process...');

        try {
            // Attempt server logout
            const response = await window.api.logout();
            console.log('✅ Server logout successful:', response);

            this.currentUser = null;
            this.isAuthenticated = false;

            // Close WebSocket
            if (window.forumApp && window.forumApp.websocket) {
                console.log('🔒 Auth: Closing WebSocket');
                window.forumApp.websocket.close();
                console.log('🔒 Auth: WebSocket closed successfully');
            }

            // Show success notification
            if (window.forumApp && window.forumApp.notificationComponent) {
                console.log('🔒 Auth: Showing success notification');
                window.forumApp.notificationComponent.info('You have been logged out.');
                console.log('🔒 Auth: Success notification completed');
            }

            // Redirect to home page
            if (window.forumApp && window.forumApp.router) {
                window.forumApp.router.navigate('/');
            }

            return { success: true };
        } catch (error) {
            console.error('❌ Logout error:', error);

            // Even if logout fails on server, clear local state for security
            this.currentUser = null;
            this.isAuthenticated = false;

            // Close WebSocket anyway
            if (window.forumApp && window.forumApp.websocket) {
                window.forumApp.websocket.close();
            }

            // Show warning notification
            if (window.forumApp && window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.warning('Logout completed locally. You have been signed out.');
            }

            // Redirect anyway for security
            if (window.forumApp && window.forumApp.router) {
                window.forumApp.router.navigate('/');
            }

            // Don't throw error - logout should always succeed from user perspective
            return { success: true, warning: 'Local logout completed' };
        }
    },

    /**
     * Check if user is authenticated
     */
    isLoggedIn() {
        return this.isAuthenticated && this.currentUser !== null;
    },

    /**
     * Get current user
     */
    getCurrentUser() {
        return this.currentUser;
    },

    /**
     * Validate registration data
     */
    validateRegistrationData(data) {
        const errors = [];

        // Email validation
        if (!data.email) {
            errors.push('Email is required');
        } else if (!window.utils.isValidEmail(data.email)) {
            errors.push('Please enter a valid email address');
        }

        // Nickname validation
        if (!data.nickname) {
            errors.push('Nickname is required');
        } else if (data.nickname.length < 3) {
            errors.push('Nickname must be at least 3 characters long');
        } else if (data.nickname.length > 20) {
            errors.push('Nickname must be less than 20 characters long');
        } else if (!/^[a-zA-Z0-9_-]+$/.test(data.nickname)) {
            errors.push('Nickname can only contain letters, numbers, underscores, and hyphens');
        }

        // Password validation
        if (!data.password) {
            errors.push('Password is required');
        } else if (!window.utils.isValidPassword(data.password)) {
            errors.push('Password must be at least 8 characters long');
        }

        // Name validation
        if (!data.firstName) {
            errors.push('First name is required');
        } else if (data.firstName.length < 2) {
            errors.push('First name must be at least 2 characters long');
        }

        if (!data.lastName) {
            errors.push('Last name is required');
        } else if (data.lastName.length < 2) {
            errors.push('Last name must be at least 2 characters long');
        }

        // Age validation
        if (!data.age) {
            errors.push('Age is required');
        } else if (data.age < 13) {
            errors.push('You must be at least 13 years old to register');
        } else if (data.age > 120) {
            errors.push('Please enter a valid age');
        }

        // Gender validation
        if (!data.gender) {
            errors.push('Gender is required');
        } else if (!['male', 'female'].includes(data.gender)) {
            errors.push('Please select a valid gender');
        }

        return {
            isValid: errors.length === 0,
            message: errors.join(', '),
            errors
        };
    },

    /**
     * Require authentication for protected routes
     */
    requireAuth() {
        if (!this.isLoggedIn()) {
            // Show notification
            if (window.forumApp && window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.warning('Please log in to access this page.');
            }
            
            // Redirect to login
            if (window.forumApp && window.forumApp.router) {
                window.forumApp.router.navigate('/login');
            }
            
            return false;
        }
        return true;
    },

    /**
     * Redirect authenticated users away from auth pages
     */
    redirectIfAuthenticated() {
        if (this.isLoggedIn()) {
            if (window.forumApp && window.forumApp.router) {
                window.forumApp.router.navigate('/');
            }
            return true;
        }
        return false;
    }
};
