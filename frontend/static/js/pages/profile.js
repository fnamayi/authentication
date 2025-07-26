// Profile Page Component
window.ProfilePage = {
    currentUser: null,
    profileStats: null,
    recentActivity: [],

    async render() {
        console.log('Profile page render called');
        window.forumApp.setCurrentPage('profile');

        // Require authentication
        if (!window.auth.requireAuth()) {
            console.log('Authentication required, redirecting');
            return;
        }

        this.currentUser = window.auth.getCurrentUser();
        console.log('Current user:', this.currentUser);

        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="profile-container">
                <!-- Profile Hero Section -->
                <div class="profile-hero">
                    <div class="profile-hero-content">
                        <div class="profile-avatar-wrapper">
                            <img id="profile-avatar"
                                 src="${this.currentUser.avatarUrl || '/static/images/default-avatar.svg'}"
                                 alt="Profile Avatar"
                                 class="profile-avatar">
                            <div class="online-indicator ${this.currentUser.isOnline ? 'online' : 'offline'}"></div>
                        </div>

                        <div class="profile-hero-info">
                            <h1 class="profile-name">${window.utils.escapeHtml(this.currentUser.nickname)}</h1>
                            <p class="profile-title">${window.utils.escapeHtml(this.currentUser.firstName)} ${window.utils.escapeHtml(this.currentUser.lastName)}</p>
                            <p class="profile-email">
                                <i class="icon-email">📧</i>
                                ${window.utils.escapeHtml(this.currentUser.email)}
                            </p>
                            <p class="profile-joined">
                                <i class="icon-calendar">📅</i>
                                Member since ${window.utils.formatDate(this.currentUser.createdAt)}
                            </p>
                        </div>

                        <div class="profile-actions">
                            <button class="btn btn-primary edit-profile-btn">
                                <i class="icon-edit">✏️</i>
                                Edit Profile
                            </button>
                        </div>
                    </div>
                </div>

            </div>
        `;

        console.log('About to bind events');
        this.bindEvents();
        console.log('Events bound');
    },

    handleTabSwitch(event) {
        const clickedTab = event.target.closest('.tab-btn');
        const tabName = clickedTab.dataset.tab;
        this.switchToTab(tabName);
    },

    switchToTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');

        // Load tab-specific data
        this.loadTabData(tabName);
    },

    async loadTabData(tabName) {
        switch (tabName) {
            case 'overview':
                await this.loadOverviewData();
                break;
            case 'activity':
                await this.loadActivityData();
                break;
            case 'posts':
                await this.loadUserPosts();
                break;
            case 'achievements':
                await this.loadAchievements();
                break;
            case 'settings':
                // Settings tab is already loaded
                break;
        }
    },

    async loadProfileData() {
        // Load all initial data
        await Promise.all([
            this.loadProfileStats(),
            this.loadOverviewData()
        ]);
    },

    async loadOverviewData() {
        await Promise.all([
            this.loadDetailedStats(),
            this.loadRecentAchievements(),
            this.loadRecentActivity()
        ]);
    },

    async loadDetailedStats() {
        const container = document.getElementById('detailed-stats');
        if (!container) return;

        try {
            // For now, we'll use mock data. In a real app, this would be an API call
            const stats = {
                postsCount: 12,
                commentsCount: 45,
                likesReceived: 89,
                friendsCount: 23,
                joinDate: this.currentUser.createdAt,
                lastActive: new Date().toISOString()
            };

            container.innerHTML = `
                <div class="stat-row">
                    <div class="stat-item">
                        <span class="stat-icon">📝</span>
                        <div class="stat-info">
                            <span class="stat-value">${stats.postsCount}</span>
                            <span class="stat-label">Posts Created</span>
                        </div>
                    </div>
                    <div class="stat-item">
                        <span class="stat-icon">💬</span>
                        <div class="stat-info">
                            <span class="stat-value">${stats.commentsCount}</span>
                            <span class="stat-label">Comments Made</span>
                        </div>
                    </div>
                </div>
                <div class="stat-row">
                    <div class="stat-item">
                        <span class="stat-icon">👍</span>
                        <div class="stat-info">
                            <span class="stat-value">${stats.likesReceived}</span>
                            <span class="stat-label">Likes Received</span>
                        </div>
                    </div>
                    <div class="stat-item">
                        <span class="stat-icon">👥</span>
                        <div class="stat-info">
                            <span class="stat-value">${stats.friendsCount}</span>
                            <span class="stat-label">Friends</span>
                        </div>
                    </div>
                </div>
                <div class="stat-row">
                    <div class="stat-item">
                        <span class="stat-icon">📅</span>
                        <div class="stat-info">
                            <span class="stat-value">${window.utils.formatDate(stats.joinDate)}</span>
                            <span class="stat-label">Member Since</span>
                        </div>
                    </div>
                    <div class="stat-item">
                        <span class="stat-icon">🕒</span>
                        <div class="stat-info">
                            <span class="stat-value">${window.utils.formatDate(stats.lastActive)}</span>
                            <span class="stat-label">Last Active</span>
                        </div>
                    </div>
                </div>
            `;

            // Update quick stats in hero section
            document.querySelector('#posts-count .stat-number').textContent = stats.postsCount;
            document.querySelector('#comments-count .stat-number').textContent = stats.commentsCount;
            document.querySelector('#likes-count .stat-number').textContent = stats.likesReceived;
            document.querySelector('#friends-count .stat-number').textContent = stats.friendsCount;

        } catch (error) {
            console.error('Failed to load detailed stats:', error);
            container.innerHTML = '<div class="error-message">Failed to load statistics</div>';
        }
    },

    async loadRecentAchievements() {
        const container = document.getElementById('achievements-preview');
        if (!container) return;

        try {
            // Mock achievements data
            const achievements = [
                { id: 1, name: 'First Post', description: 'Created your first post', icon: '🎉', earned: true },
                { id: 2, name: 'Social Butterfly', description: 'Made 10 friends', icon: '🦋', earned: true },
                { id: 3, name: 'Conversation Starter', description: 'Received 50 likes', icon: '💬', earned: false }
            ];

            container.innerHTML = achievements.map(achievement => `
                <div class="achievement-item ${achievement.earned ? 'earned' : 'locked'}">
                    <span class="achievement-icon">${achievement.icon}</span>
                    <div class="achievement-info">
                        <h4 class="achievement-name">${achievement.name}</h4>
                        <p class="achievement-desc">${achievement.description}</p>
                    </div>
                    ${achievement.earned ? '<span class="achievement-badge">✓</span>' : '<span class="achievement-lock">🔒</span>'}
                </div>
            `).join('');

        } catch (error) {
            console.error('Failed to load achievements:', error);
            container.innerHTML = '<div class="error-message">Failed to load achievements</div>';
        }
    },

    async loadRecentActivity() {
        const container = document.getElementById('activity-preview');
        if (!container) return;

        try {
            // Mock activity data
            const activities = [
                { type: 'post', action: 'Created a new post', title: 'My thoughts on technology', time: new Date(Date.now() - 2 * 60 * 60 * 1000) },
                { type: 'comment', action: 'Commented on', title: 'Discussion about AI', time: new Date(Date.now() - 5 * 60 * 60 * 1000) },
                { type: 'like', action: 'Liked a post', title: 'Best programming practices', time: new Date(Date.now() - 8 * 60 * 60 * 1000) }
            ];

            container.innerHTML = activities.map(activity => `
                <div class="activity-item">
                    <div class="activity-icon ${activity.type}">
                        ${activity.type === 'post' ? '📝' : activity.type === 'comment' ? '💬' : '👍'}
                    </div>
                    <div class="activity-content">
                        <p class="activity-text">
                            ${activity.action} <strong>${activity.title}</strong>
                        </p>
                        <span class="activity-time">${window.utils.formatDate(activity.time)}</span>
                    </div>
                </div>
            `).join('');

        } catch (error) {
            console.error('Failed to load recent activity:', error);
            container.innerHTML = '<div class="error-message">Failed to load activity</div>';
        }
    },

    bindEvents() {
        console.log('bindEvents called');

        // Edit profile button
        const editProfileBtn = document.querySelector('.edit-profile-btn');
        if (editProfileBtn) {
            editProfileBtn.addEventListener('click', () => {
                this.showEditProfileModal();
            });
        }
    },

    showEditProfileModal() {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal-content profile-edit-modal">
                <div class="modal-header">
                    <h3>Edit Profile</h3>
                    <button class="modal-close">&times;</button>
                </div>

                <form id="edit-profile-form" class="modal-form">
                    <!-- Profile Picture Section -->
                    <div class="form-group profile-picture-section">
                        <label>Profile Picture</label>
                        <div class="profile-picture-upload">
                            <div class="current-avatar">
                                <img id="modal-avatar-preview"
                                     src="${this.currentUser.avatarUrl || '/static/images/default-avatar.svg'}"
                                     alt="Current Avatar"
                                     class="avatar-preview">
                            </div>
                            <div class="upload-controls">
                                <input type="file" id="avatar-file-input" accept="image/*" style="display: none;">
                                <button type="button" class="btn btn-outline upload-avatar-btn">
                                    📷 Choose New Picture
                                </button>
                                <button type="button" class="btn btn-outline remove-avatar-btn"
                                        style="display: ${this.currentUser.avatarUrl ? 'inline-block' : 'none'};">
                                    🗑️ Remove Picture
                                </button>
                            </div>
                            <small class="form-help">Supported formats: JPG, PNG, GIF (max 5MB)</small>
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="edit-firstName">First Name</label>
                        <input type="text" id="edit-firstName" name="firstName" required
                               value="${window.utils.escapeHtml(this.currentUser.firstName)}">
                    </div>

                    <div class="form-group">
                        <label for="edit-lastName">Last Name</label>
                        <input type="text" id="edit-lastName" name="lastName" required
                               value="${window.utils.escapeHtml(this.currentUser.lastName)}">
                    </div>

                    <div class="form-group">
                        <label for="edit-age">Age</label>
                        <input type="number" id="edit-age" name="age" min="13" max="120" required
                               value="${this.currentUser.age || ''}">
                    </div>

                    <div class="form-group">
                        <label for="edit-gender">Gender</label>
                        <select id="edit-gender" name="gender" required>
                            <option value="">Select Gender</option>
                            <option value="male" ${this.currentUser.gender === 'male' ? 'selected' : ''}>Male</option>
                            <option value="female" ${this.currentUser.gender === 'female' ? 'selected' : ''}>Female</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="edit-email">Email</label>
                        <input type="email" id="edit-email" name="email" disabled
                               value="${window.utils.escapeHtml(this.currentUser.email)}">
                        <small class="form-help">Email cannot be changed</small>
                    </div>

                    <div class="modal-actions">
                        <button type="button" class="btn btn-outline modal-cancel">Cancel</button>
                        <button type="submit" class="btn btn-primary">Update Profile</button>
                    </div>
                </form>
            </div>
        `;

        document.body.appendChild(modal);

        // Bind modal events
        const closeBtn = modal.querySelector('.modal-close');
        const cancelBtn = modal.querySelector('.modal-cancel');
        const form = modal.querySelector('#edit-profile-form');
        const avatarFileInput = modal.querySelector('#avatar-file-input');
        const uploadAvatarBtn = modal.querySelector('.upload-avatar-btn');
        const removeAvatarBtn = modal.querySelector('.remove-avatar-btn');
        const avatarPreview = modal.querySelector('#modal-avatar-preview');

        let selectedAvatarFile = null;
        let avatarRemoved = false;

        const closeModal = () => {
            document.body.removeChild(modal);
        };

        closeBtn.addEventListener('click', closeModal);
        cancelBtn.addEventListener('click', closeModal);
        modal.addEventListener('click', (e) => {
            if (e.target === modal) closeModal();
        });

        // Avatar upload functionality
        uploadAvatarBtn.addEventListener('click', () => {
            avatarFileInput.click();
        });

        avatarFileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                // Validate file type
                if (!file.type.startsWith('image/')) {
                    if (window.forumApp.notificationComponent) {
                        window.forumApp.notificationComponent.error('Please select an image file');
                    }
                    return;
                }

                // Validate file size (5MB limit)
                if (file.size > 5 * 1024 * 1024) {
                    if (window.forumApp.notificationComponent) {
                        window.forumApp.notificationComponent.error('Image size must be less than 5MB');
                    }
                    return;
                }

                selectedAvatarFile = file;
                avatarRemoved = false;

                // Preview the selected image
                const reader = new FileReader();
                reader.onload = (e) => {
                    avatarPreview.src = e.target.result;
                    removeAvatarBtn.style.display = 'inline-block';
                };
                reader.readAsDataURL(file);
            }
        });

        removeAvatarBtn.addEventListener('click', () => {
            selectedAvatarFile = null;
            avatarRemoved = true;
            avatarPreview.src = '/static/images/default-avatar.svg';
            removeAvatarBtn.style.display = 'none';
            avatarFileInput.value = '';
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const formData = new FormData(form);
            const profileData = {
                firstName: formData.get('firstName').trim(),
                lastName: formData.get('lastName').trim(),
                age: parseInt(formData.get('age')),
                gender: formData.get('gender')
            };

            const submitBtn = form.querySelector('button[type="submit"]');

            try {
                window.utils.setLoading(submitBtn, true, 'Updating...');

                // Step 1: Handle avatar upload/removal if needed
                if (selectedAvatarFile) {
                    console.log('📷 Uploading new avatar...');
                    const avatarFormData = new FormData();
                    avatarFormData.append('avatar', selectedAvatarFile);

                    const avatarResponse = await window.api.uploadAvatar(avatarFormData);
                    if (avatarResponse.success) {
                        this.currentUser.avatarUrl = avatarResponse.data.avatarUrl;
                        console.log('✅ Avatar uploaded successfully');
                    } else {
                        throw new Error(avatarResponse.message || 'Failed to upload avatar');
                    }
                } else if (avatarRemoved && this.currentUser.avatarUrl) {
                    console.log('🗑️ Removing avatar...');
                    const avatarResponse = await window.api.updateAvatar({ avatarUrl: null });
                    if (avatarResponse.success) {
                        this.currentUser.avatarUrl = null;
                        console.log('✅ Avatar removed successfully');
                    } else {
                        throw new Error(avatarResponse.message || 'Failed to remove avatar');
                    }
                }

                // Step 2: Update profile data
                console.log('📝 Updating profile data...');
                const profileResponse = await window.api.updateProfile(profileData);
                if (profileResponse.success) {
                    // Update local user data
                    this.currentUser.firstName = profileData.firstName;
                    this.currentUser.lastName = profileData.lastName;
                    this.currentUser.age = profileData.age;
                    this.currentUser.gender = profileData.gender;

                    // Update the display
                    document.querySelector('.profile-title').textContent =
                        `${this.currentUser.firstName} ${this.currentUser.lastName}`;

                    // Update avatar if changed
                    const profileAvatar = document.querySelector('#profile-avatar');
                    if (profileAvatar) {
                        profileAvatar.src = this.currentUser.avatarUrl || '/static/images/default-avatar.svg';
                    }

                    console.log('✅ Profile updated successfully');

                    if (window.forumApp.notificationComponent) {
                        window.forumApp.notificationComponent.success('Profile updated successfully');
                    }

                    closeModal();
                } else {
                    throw new Error(profileResponse.message || 'Failed to update profile');
                }

            } catch (error) {
                console.error('❌ Failed to update profile:', error);
                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.error(error.message || 'Failed to update profile');
                }
            } finally {
                window.utils.setLoading(submitBtn, false);
            }
        });

        // Focus first input
        modal.querySelector('#edit-firstName').focus();
    }









};
