// Home Page Component
window.HomePage = {
    async render() {
        window.forumApp.setCurrentPage('home');

        const mainContent = document.getElementById('main-content');
        
        if (window.auth.isLoggedIn()) {
            // Authenticated home page
            mainContent.innerHTML = `
                <div class="home-container">
                    <div class="welcome-section">
                        <h1>Welcome back, ${window.utils.escapeHtml(window.auth.getCurrentUser().nickname)}!</h1>
                        <p>What would you like to do today?</p>
                    </div>
                    
                    <div class="quick-actions">
                        <div class="action-card">
                            <h3>📝 Create a Post</h3>
                            <p>Share your thoughts with the community</p>
                            <a href="/create-post" data-route="/create-post" class="btn btn-primary">Create Post</a>
                        </div>
                        
                        <div class="action-card">
                            <h3>💬 Browse Posts</h3>
                            <p>See what others are discussing</p>
                            <a href="/posts" data-route="/posts" class="btn btn-secondary">View Posts</a>
                        </div>
                        
                        <div class="action-card">
                            <h3>📨 Messages</h3>
                            <p>Chat with other users privately</p>
                            <a href="/messages" data-route="/messages" class="btn btn-secondary">Open Messages</a>
                        </div>
                        
                       
                    </div>
                    
                    <div class="recent-activity">
                        <h2>Recent Activity</h2>
                        <div id="recent-posts" class="recent-posts-container">
                            <div class="loading-placeholder">Loading recent posts...</div>
                        </div>
                    </div>
                </div>
            `;
            
            // Load recent posts
            this.loadRecentPosts();
            
        } else {
            // Unauthenticated home page
            mainContent.innerHTML = `
                <div class="home-container">
                    <div class="hero-section">
                        <h1>🌐 Welcome to Forum</h1>
                        <p class="hero-subtitle">A modern real-time discussion platform</p>
                        <p class="hero-description">
                            Join our community to share ideas, connect with others, and participate in meaningful discussions.
                            Experience real-time messaging, category-based posts, and a vibrant community atmosphere.
                        </p>
                        
                        <div class="hero-actions">
                            <a href="/register" data-route="/register" class="btn btn-primary btn-large">
                                Get Started
                            </a>
                            <a href="/login" data-route="/login" class="btn btn-outline btn-large">
                                Sign In
                            </a>
                        </div>
                    </div>
                    
                    <div class="features-section">
                        <h2>Why Choose Forum?</h2>
                        <div class="features-grid">
                            <div class="feature-card">
                                <div class="feature-icon">⚡</div>
                                <h3>Real-time Communication</h3>
                                <p>Instant messaging and live updates keep you connected with the community</p>
                            </div>
                            
                            <div class="feature-card">
                                <div class="feature-icon">📂</div>
                                <h3>Organized Discussions</h3>
                                <p>Category-based posts help you find and participate in topics you care about</p>
                            </div>
                            
                            <div class="feature-card">
                                <div class="feature-icon">👥</div>
                                <h3>Community Focused</h3>
                                <p>Build friendships, follow interesting users, and grow your network</p>
                            </div>
                            
                            <div class="feature-card">
                                <div class="feature-icon">🔒</div>
                                <h3>Secure & Private</h3>
                                <p>Your data is protected with modern security practices and privacy controls</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="cta-section">
                        <h2>Ready to Join?</h2>
                        <p>Create your account today and become part of our growing community</p>
                        <a href="/register" data-route="/register" class="btn btn-primary btn-large">
                            Join Forum Now
                        </a>
                    </div>
                </div>
            `;
        }
    },

    async loadRecentPosts() {
        try {
            const response = await window.api.getPosts({ limit: 5 });
            if (response.success) {
                this.renderRecentPosts(response.data || []);
            }
        } catch (error) {
            console.error('Failed to load recent posts:', error);
            const container = document.getElementById('recent-posts');
            if (container) {
                container.innerHTML = '<p class="error-message">Failed to load recent posts</p>';
            }
        }
    },

    renderRecentPosts(posts) {
        const container = document.getElementById('recent-posts');
        if (!container) return;

        if (posts.length === 0) {
            container.innerHTML = '<p class="no-posts">No posts yet. Be the first to create one!</p>';
            return;
        }

        container.innerHTML = posts.map(post => `
            <div class="recent-post" data-post-id="${post.id}">
                <div class="post-header">
                    <h4>${window.utils.escapeHtml(post.title)}</h4>
                    <span class="post-meta">by ${window.utils.escapeHtml(post.author)} • ${window.utils.formatDate(post.createdAt)}</span>
                </div>
                <div class="post-excerpt">
                    ${window.utils.escapeHtml(post.content.substring(0, 150))}${post.content.length > 150 ? '...' : ''}
                </div>
                <div class="post-categories">
                    ${post.categories.map(cat => `<span class="category-tag">${window.utils.escapeHtml(cat)}</span>`).join('')}
                </div>
                <div class="post-stats">
                    <span class="stat-display">👍 ${post.likeCount}</span>
                    <span class="stat-display">👎 ${post.dislikeCount}</span>
                    <span class="stat-display">💬 ${post.commentCount}</span>
                    <span class="click-to-interact">Click to interact →</span>
                </div>
            </div>
        `).join('');

        // Bind click events to navigate to posts page
        this.bindRecentPostEvents();
    },

    bindRecentPostEvents() {
        const recentPosts = document.querySelectorAll('.recent-post');
        recentPosts.forEach(postElement => {
            postElement.addEventListener('click', (event) => {
                // Don't navigate if clicking on category tags
                if (event.target.classList.contains('category-tag')) {
                    return;
                }

                const postId = postElement.dataset.postId;
                console.log('📱 Navigating to post:', postId);

                // Navigate to posts page with the specific post highlighted
                window.forumApp.router.navigate('/posts', {
                    highlightPost: postId,
                    scrollToPost: postId
                });
            });

            // Add hover effect
            postElement.style.cursor = 'pointer';
        });
    }
};
