// Register Page Component
window.RegisterPage = {
    async render() {
        window.forumApp.setCurrentPage('register');
        
        // Redirect if already authenticated
        if (window.auth.redirectIfAuthenticated()) {
            return;
        }

        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="auth-container">
                <div class="auth-card">
                    <div class="auth-header">
                        <h1>Join Forum</h1>
                        <p>Create your account to start participating</p>
                    </div>
                    
                    <form id="register-form" class="auth-form">
                        <!-- Error message area -->
                        <div id="register-error" class="error-message" style="display: none;"></div>

                        <div class="form-row">
                            <div class="form-group">
                                <label for="firstName">First Name</label>
                                <input 
                                    type="text" 
                                    id="firstName" 
                                    name="firstName" 
                                    required 
                                    placeholder="First name"
                                    autocomplete="given-name"
                                >
                            </div>
                            <div class="form-group">
                                <label for="lastName">Last Name</label>
                                <input 
                                    type="text" 
                                    id="lastName" 
                                    name="lastName" 
                                    required 
                                    placeholder="Last name"
                                    autocomplete="family-name"
                                >
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label for="email">Email</label>
                            <input 
                                type="email" 
                                id="email" 
                                name="email" 
                                required 
                                placeholder="Enter your email"
                                autocomplete="email"
                            >
                        </div>
                        
                        <div class="form-group">
                            <label for="nickname">Nickname</label>
                            <input 
                                type="text" 
                                id="nickname" 
                                name="nickname" 
                                required 
                                placeholder="Choose a unique nickname"
                                autocomplete="username"
                                pattern="[a-zA-Z0-9_-]+"
                                title="Only letters, numbers, underscores, and hyphens allowed"
                            >
                        </div>
                        
                        <div class="form-row">
                            <div class="form-group">
                                <label for="age">Age</label>
                                <input 
                                    type="number" 
                                    id="age" 
                                    name="age" 
                                    required 
                                    min="13" 
                                    max="120"
                                    placeholder="Age"
                                >
                            </div>
                            <div class="form-group">
                                <label for="gender">Gender</label>
                                <select id="gender" name="gender" required>
                                    <option value="">Select gender</option>
                                    <option value="male">Male</option>
                                    <option value="female">Female</option>
                                </select>
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label for="password">Password</label>
                            <input 
                                type="password" 
                                id="password" 
                                name="password" 
                                required 
                                placeholder="Create a strong password"
                                autocomplete="new-password"
                                minlength="8"
                            >
                            <small class="form-help">Password must be at least 8 characters long</small>
                        </div>
                        
                        <button type="submit" class="btn btn-primary btn-full">
                            Create Account
                        </button>
                    </form>
                    
                    <div class="auth-divider">
                        <span>or</span>
                    </div>
                    
                    <div class="oauth-buttons">
                        <a href="/auth/google/login" class="btn btn-outline btn-full oauth-btn">
                            <span class="oauth-icon">🔍</span>
                            Continue with Google
                        </a>
                        <a href="/auth/github/login" class="btn btn-outline btn-full oauth-btn">
                            <span class="oauth-icon">🐙</span>
                            Continue with GitHub
                        </a>
                    </div>
                    
                    <div class="auth-footer">
                        <p>Already have an account? <a href="/login" data-route="/login">Sign in</a></p>
                    </div>
                </div>
            </div>
        `;

        this.bindEvents();
    },

    bindEvents() {
        const registerForm = document.getElementById('register-form');
        if (registerForm) {
            registerForm.addEventListener('submit', this.handleRegister.bind(this));
        }

        // Real-time validation
        const nicknameInput = document.getElementById('nickname');
        if (nicknameInput) {
            nicknameInput.addEventListener('input', this.validateNickname.bind(this));
        }

        const emailInput = document.getElementById('email');
        if (emailInput) {
            emailInput.addEventListener('input', this.validateEmail.bind(this));
        }
    },

    validateNickname(event) {
        const input = event.target;
        const value = input.value;
        
        // Remove invalid characters
        const sanitized = value.replace(/[^a-zA-Z0-9_-]/g, '');
        if (sanitized !== value) {
            input.value = sanitized;
        }
    },

    validateEmail(event) {
        const input = event.target;
        const value = input.value;
        
        if (value && !window.utils.isValidEmail(value)) {
            input.setCustomValidity('Please enter a valid email address');
        } else {
            input.setCustomValidity('');
        }
    },

    async handleRegister(event) {
        event.preventDefault();

        const form = event.target;
        const formData = new FormData(form);

        const userData = {
            firstName: formData.get('firstName').trim(),
            lastName: formData.get('lastName').trim(),
            email: formData.get('email').trim(),
            nickname: formData.get('nickname').trim(),
            age: parseInt(formData.get('age')),
            gender: formData.get('gender'),
            password: formData.get('password')
        };

        const submitBtn = form.querySelector('button[type="submit"]');

        // Clear any previous errors
        this.hideError();

        // Basic validation
        if (!userData.firstName || !userData.lastName || !userData.email || !userData.nickname || !userData.password) {
            this.showError('Please fill in all required fields');
            return;
        }

        if (userData.age < 13) {
            this.showError('You must be at least 13 years old to register');
            return;
        }

        try {
            // Show loading state
            window.utils.setLoading(submitBtn, true, 'Creating Account...');

            // Attempt registration
            const result = await window.auth.register(userData);

            if (result.success) {
                // Update app state
                window.forumApp.onAuthSuccess(result.user);
            } else {
                // Show error from server response
                this.showError(result.message || 'Registration failed. Please check your information and try again.');
            }

        } catch (error) {
            console.error('Registration error:', error);

            // Show error in form
            this.showError(error.message || 'Registration failed. Please try again.');

            // Also show notification as backup
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error(error.message || 'Registration failed');
            }

        } finally {
            // Remove loading state
            window.utils.setLoading(submitBtn, false);
        }
    },

    showError(message) {
        const errorDiv = document.getElementById('register-error');
        if (errorDiv) {
            errorDiv.textContent = message;
            errorDiv.style.display = 'block';

            // Scroll error into view
            errorDiv.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    },

    hideError() {
        const errorDiv = document.getElementById('register-error');
        if (errorDiv) {
            errorDiv.style.display = 'none';
            errorDiv.textContent = '';
        }
    }
};
