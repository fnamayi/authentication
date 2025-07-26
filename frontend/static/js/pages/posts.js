// Posts Page Component
window.PostsPage = {
    async render(options = {}) {
        window.forumApp.setCurrentPage('posts');

        // Store options for later use
        this.highlightPostId = options.highlightPost;
        this.scrollToPostId = options.scrollToPost;

        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="posts-container">
                <div class="posts-header">
                    <h1>Community Posts</h1>
                    <div class="posts-actions">
                        ${window.auth.isLoggedIn() ?
                            '<a href="/create-post" data-route="/create-post" class="btn btn-primary">Create Post</a>' :
                            '<a href="/login" data-route="/login" class="btn btn-primary">Login to Post</a>'
                        }
                    </div>
                </div>

                <div class="posts-filters">
                    <div class="filter-group">
                        <label for="category-filter">Filter by Category:</label>
                        <select id="category-filter">
                            <option value="">All Categories</option>
                        </select>
                    </div>
                </div>

                <div id="posts-list" class="posts-list">
                    <div class="loading-placeholder">Loading posts...</div>
                </div>
            </div>
        `;

        await this.loadCategories();
        await this.loadPosts();
        this.bindEvents();

        // Handle post highlighting and scrolling
        if (this.highlightPostId || this.scrollToPostId) {
            setTimeout(() => {
                this.handlePostNavigation();
            }, 500); // Wait for posts to render
        }
    },

    bindEvents() {
        const categoryFilter = document.getElementById('category-filter');
        if (categoryFilter) {
            categoryFilter.addEventListener('change', () => {
                this.loadPosts(categoryFilter.value);
            });
        }
    },

    async loadCategories() {
        try {
            const response = await window.api.getCategories();
            if (response.success) {
                this.renderCategories(response.data || []);
            }
        } catch (error) {
            console.error('Failed to load categories:', error);
        }
    },

    renderCategories(categories) {
        const categoryFilter = document.getElementById('category-filter');
        if (!categoryFilter) return;

        // Keep the "All Categories" option and add others
        const currentValue = categoryFilter.value;
        categoryFilter.innerHTML = '<option value="">All Categories</option>';

        // Add predefined categories
        const predefinedCategories = [
            'General', 'Technology', 'Sports', 'Entertainment', 'Gaming',
            'Music', 'Movies', 'Food', 'Travel', 'Health', 'Education', 'Science'
        ];

        predefinedCategories.forEach(categoryName => {
            const category = categories.find(c => c.name === categoryName);
            const option = document.createElement('option');
            option.value = categoryName;
            option.textContent = category ? `${categoryName} (${category.postCount})` : categoryName;
            categoryFilter.appendChild(option);
        });

        // Add any other categories that exist in the database but aren't in predefined list
        categories.forEach(category => {
            if (!predefinedCategories.includes(category.name)) {
                const option = document.createElement('option');
                option.value = category.name;
                option.textContent = `${category.name} (${category.postCount})`;
                categoryFilter.appendChild(option);
            }
        });

        categoryFilter.value = currentValue;
    },

    async loadPosts(category = '') {
        try {
            const params = category ? { category } : {};
            const response = await window.api.getPosts(params);
            if (response.success) {
                this.renderPosts(response.data || []);
            }
        } catch (error) {
            console.error('Failed to load posts:', error);
            const container = document.getElementById('posts-list');
            if (container) {
                container.innerHTML = '<div class="error-message">Failed to load posts</div>';
            }
        }
    },

    renderPosts(posts) {
        const container = document.getElementById('posts-list');
        if (!container) return;

        if (posts.length === 0) {
            container.innerHTML = '<div class="no-posts">No posts found. Be the first to create one!</div>';
            return;
        }

        container.innerHTML = posts.map(post => `
            <article class="post-card">
                <div class="post-header">
                    <div class="post-author">
                        <img src="${post.authorAvatar || '/static/images/default-avatar.png'}" 
                             alt="${window.utils.escapeHtml(post.author)}'s avatar" 
                             class="author-avatar">
                        <div class="author-info">
                            <span class="author-name">${window.utils.escapeHtml(post.author)}</span>
                            <span class="post-date">${window.utils.formatDate(post.createdAt)}</span>
                        </div>
                    </div>
                </div>
                
                <div class="post-content">
                    <h2 class="post-title">
                        <a href="/post/${post.id}" data-route="/post/${post.id}">
                            ${window.utils.escapeHtml(post.title)}
                        </a>
                    </h2>
                    
                    <div class="post-excerpt">
                        ${window.utils.escapeHtml(post.content.substring(0, 200))}${post.content.length > 200 ? '...' : ''}
                    </div>
                    
                    ${post.categories.length > 0 ? `
                        <div class="post-categories">
                            ${post.categories.map(cat => `<span class="category-tag">${window.utils.escapeHtml(cat)}</span>`).join('')}
                        </div>
                    ` : ''}
                </div>
                
                <div class="post-footer">
                    <div class="post-stats">
                        <button class="stat-btn like-btn ${post.userLiked ? 'active' : ''}" 
                                data-post-id="${post.id}" data-action="like">
                            👍 ${post.likeCount}
                        </button>
                        <button class="stat-btn dislike-btn ${post.userDisliked ? 'active' : ''}" 
                                data-post-id="${post.id}" data-action="dislike">
                            👎 ${post.dislikeCount}
                        </button>
                        <button class="stat-btn comment-toggle-btn"
                                data-post-id="${post.id}"
                                title="Toggle comments">
                            💬 ${post.commentCount}
                        </button>
                    </div>

                </div>
            </article>

            <!-- Comments section as separate container -->
            <div class="post-comments-container" id="comments-container-${post.id}">
                <div class="comments-section" id="comments-${post.id}" style="display: none;">
                    <div class="comments-loading">Loading comments...</div>
                </div>
            </div>
        `).join('');

        // Bind like/dislike events
        this.bindPostEvents();
    },

    bindPostEvents() {
        const likeButtons = document.querySelectorAll('.like-btn, .dislike-btn');
        likeButtons.forEach(button => {
            button.addEventListener('click', this.handleLikeDislike.bind(this));
        });

        // Bind comment toggle buttons
        const commentToggleButtons = document.querySelectorAll('.comment-toggle-btn');
        commentToggleButtons.forEach(button => {
            button.addEventListener('click', this.handleCommentToggle.bind(this));
        });
    },

    async handleLikeDislike(event) {
        if (!window.auth.isLoggedIn()) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.warning('Please log in to like posts');
            }
            return;
        }

        const button = event.target;
        const postId = parseInt(button.dataset.postId);
        const isLike = button.dataset.action === 'like';

        try {
            const response = await window.api.likePost(postId, isLike);

            if (response.success && response.data) {
                // Update the UI with new counts
                const postCard = button.closest('.post-card');
                const likeBtn = postCard.querySelector('.like-btn');
                const dislikeBtn = postCard.querySelector('.dislike-btn');

                if (likeBtn) {
                    likeBtn.innerHTML = `👍 ${response.data.likeCount}`;
                    likeBtn.classList.toggle('active', isLike);
                }
                if (dislikeBtn) {
                    dislikeBtn.innerHTML = `👎 ${response.data.dislikeCount}`;
                    dislikeBtn.classList.toggle('active', !isLike);
                }
            }

        } catch (error) {
            console.error('Failed to like/dislike post:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to update like status');
            }
        }
    },

    async handleCommentToggle(event) {
        event.preventDefault();
        event.stopPropagation();

        const button = event.currentTarget;
        const postId = button.dataset.postId;
        const commentsSection = document.getElementById(`comments-${postId}`);

        if (!commentsSection) {
            console.error('Comments section not found for post:', postId);
            return;
        }

        // Toggle visibility
        const isVisible = commentsSection.style.display !== 'none';

        if (isVisible) {
            // Hide comments with animation
            commentsSection.style.opacity = '0';
            commentsSection.style.transform = 'translateY(-10px)';
            setTimeout(() => {
                commentsSection.style.display = 'none';
            }, 200);
            button.classList.remove('active');
        } else {
            // Show comments and load them if not already loaded
            commentsSection.style.display = 'block';
            commentsSection.style.opacity = '0';
            commentsSection.style.transform = 'translateY(-10px)';
            button.classList.add('active');

            // Load comments if not already loaded
            if (!commentsSection.dataset.loaded) {
                await this.loadCommentsForPost(postId);
                commentsSection.dataset.loaded = 'true';
            }

            // Animate in
            setTimeout(() => {
                commentsSection.style.opacity = '1';
                commentsSection.style.transform = 'translateY(0)';

                // Smooth scroll to comments section
                commentsSection.scrollIntoView({
                    behavior: 'smooth',
                    block: 'nearest'
                });
            }, 50);
        }
    },

    async loadCommentsForPost(postId) {
        const commentsSection = document.getElementById(`comments-${postId}`);
        if (!commentsSection) return;

        try {
            console.log('🔄 Loading comments for post:', postId);

            // Show loading state
            commentsSection.innerHTML = '<div class="comments-loading">Loading comments...</div>';

            // Fetch comments from API
            const response = await window.api.getComments(postId);
            console.log('📡 Comments API response:', response);
            console.log('📡 Response success:', response.success);
            console.log('📡 Response data:', response.data);
            console.log('📡 Response data type:', typeof response.data);
            console.log('📡 Response data is array:', Array.isArray(response.data));

            if (response.success) {
                // Handle both cases: response.data exists or is null/undefined
                const comments = response.data || [];
                console.log('✅ Loaded comments:', comments);
                console.log('✅ Comments length:', comments.length);

                if (comments.length === 0) {
                    commentsSection.innerHTML = `
                        <div class="no-comments">
                            <p>No comments yet. Be the first to comment!</p>
                        </div>
                        ${this.renderCommentForm(postId)}
                    `;
                } else {
                    commentsSection.innerHTML = `
                        <div class="comments-header">
                            <h4>Comments (${comments.length})</h4>
                        </div>
                        <div class="comments-list">
                            ${comments.map(comment => this.renderComment(comment)).join('')}
                        </div>
                        ${this.renderCommentForm(postId)}
                    `;
                }

                // Bind comment form events after rendering
                this.bindCommentFormEvents();
            } else {
                console.error('❌ API response failed:', response);
                // Even on API failure, show the comment form for posts that exist
                commentsSection.innerHTML = `
                    <div class="comments-error">
                        <p>Unable to load comments at the moment.</p>
                    </div>
                    ${this.renderCommentForm(postId)}
                `;
                this.bindCommentFormEvents();
            }
        } catch (error) {
            console.error('❌ Failed to load comments:', error);
            // Show error but still provide comment form
            commentsSection.innerHTML = `
                <div class="comments-error">
                    <p>Unable to load comments at the moment.</p>
                </div>
                ${this.renderCommentForm(postId)}
            `;
            this.bindCommentFormEvents();
        }
    },

    renderComment(comment) {
        const timeAgo = window.utils.formatDate(comment.createdAt);
        const isOwnComment = window.forumApp.currentUser && comment.userId === window.forumApp.currentUser.id;

        return `
            <div class="comment-item" data-comment-id="${comment.id}">
                <div class="comment-header">
                    <img src="${comment.authorAvatar || '/static/images/default-avatar.png'}"
                         alt="${window.utils.escapeHtml(comment.author)}'s avatar"
                         class="comment-avatar">
                    <div class="comment-meta">
                        <span class="comment-author">${window.utils.escapeHtml(comment.author)}</span>
                        <span class="comment-time">${timeAgo}</span>
                    </div>
                </div>
                <div class="comment-content" data-original-content="${window.utils.escapeHtml(comment.content)}">
                    ${window.utils.escapeHtml(comment.content)}
                </div>
                <div class="comment-actions">
                    <button class="comment-action-btn like-comment-btn ${comment.userLiked ? 'active' : ''}"
                            data-comment-id="${comment.id}" data-action="like">
                        👍 ${comment.likeCount || 0}
                    </button>
                    <button class="comment-action-btn dislike-comment-btn ${comment.userDisliked ? 'active' : ''}"
                            data-comment-id="${comment.id}" data-action="dislike">
                        👎 ${comment.dislikeCount || 0}
                    </button>
                    ${isOwnComment ? `
                        <button class="comment-action-btn edit-comment-btn"
                                data-comment-id="${comment.id}">
                            ✏️ Edit
                        </button>
                        <button class="comment-action-btn delete-comment-btn"
                                data-comment-id="${comment.id}">
                            🗑️ Delete
                        </button>
                    ` : ''}
                </div>
            </div>
        `;
    },

    renderCommentForm(postId) {
        // Check if user is logged in
        if (!window.forumApp.currentUser) {
            return `
                <div class="comment-form-container">
                    <div class="login-prompt">
                        <p>Please <a href="/login" data-route="/login">login</a> to add a comment.</p>
                    </div>
                </div>
            `;
        }

        return `
            <div class="comment-form-container">
                <form class="comment-form" data-post-id="${postId}">
                    <div class="comment-input-group">
                        <img src="${window.forumApp.currentUser.avatarUrl || '/static/images/default-avatar.png'}"
                             alt="Your avatar"
                             class="comment-form-avatar">
                        <textarea
                            class="comment-textarea"
                            placeholder="Write a comment..."
                            rows="2"
                            maxlength="500"
                            required></textarea>
                    </div>
                    <div class="comment-form-actions">
                        <span class="comment-char-count">0/500</span>
                        <button type="submit" class="btn btn-primary btn-sm">Post Comment</button>
                    </div>
                </form>
            </div>
        `;
    },

    async handleCommentSubmit(event) {
        event.preventDefault();

        const form = event.target;
        const postId = form.dataset.postId;
        const textarea = form.querySelector('.comment-textarea');
        const submitBtn = form.querySelector('button[type="submit"]');
        const content = textarea.value.trim();

        if (!content) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Please enter a comment');
            }
            return;
        }

        try {
            // Show loading state
            window.utils.setLoading(submitBtn, true, 'Posting...');
            textarea.disabled = true;

            // Submit comment
            const response = await window.api.createComment({
                postId: parseInt(postId),
                content: content
            });

            if (response.success) {
                // Clear form
                textarea.value = '';

                // Update comment count in the toggle button BEFORE reloading comments
                const toggleBtn = document.querySelector(`[data-post-id="${postId}"].comment-toggle-btn`);
                if (toggleBtn) {
                    const currentCount = parseInt(toggleBtn.textContent.match(/\d+/)[0]) || 0;
                    const newCount = currentCount + 1;
                    toggleBtn.innerHTML = `💬 ${newCount}`;
                    console.log(`📊 Updated comment count from ${currentCount} to ${newCount} for post ${postId}`);
                } else {
                    console.log(`❌ Could not find toggle button for post ${postId}`);
                }

                // Reload comments to show the new comment
                await this.loadCommentsForPost(postId);

                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success('Comment posted successfully!');
                }
            } else {
                throw new Error(response.message || 'Failed to post comment');
            }
        } catch (error) {
            console.error('Failed to post comment:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error(error.message || 'Failed to post comment');
            }
        } finally {
            // Reset loading state
            window.utils.setLoading(submitBtn, false);
            textarea.disabled = false;
            textarea.focus();
        }
    },



    bindCommentFormEvents() {
        // Bind comment form submissions
        const commentForms = document.querySelectorAll('.comment-form');
        commentForms.forEach(form => {
            form.addEventListener('submit', this.handleCommentSubmit.bind(this));
        });

        // Auto-resize textarea
        const textareas = document.querySelectorAll('.comment-textarea');
        textareas.forEach(textarea => {
            textarea.addEventListener('input', () => {
                textarea.style.height = 'auto';
                textarea.style.height = Math.min(textarea.scrollHeight, 80) + 'px';
            });
        });

        // Bind comment action buttons
        this.bindCommentActions();
    },

    bindCommentActions() {
        // Like/Dislike buttons
        const likeDislikeButtons = document.querySelectorAll('.like-comment-btn, .dislike-comment-btn');
        likeDislikeButtons.forEach(button => {
            button.addEventListener('click', this.handleCommentLikeDislike.bind(this));
        });

        // Edit buttons
        const editButtons = document.querySelectorAll('.edit-comment-btn');
        editButtons.forEach(button => {
            button.addEventListener('click', this.handleCommentEdit.bind(this));
        });

        // Delete buttons
        const deleteButtons = document.querySelectorAll('.delete-comment-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', this.handleCommentDelete.bind(this));
        });
    },

    async handleCommentLikeDislike(event) {
        event.preventDefault();
        event.stopPropagation();

        const button = event.currentTarget;
        const commentId = button.dataset.commentId;
        const action = button.dataset.action;
        const isLike = action === 'like';

        if (!window.forumApp.currentUser) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Please login to like comments');
            }
            return;
        }

        // Get the comment item and both buttons
        const commentItem = button.closest('.comment-item');
        const likeBtn = commentItem.querySelector('.like-comment-btn');
        const dislikeBtn = commentItem.querySelector('.dislike-comment-btn');

        // Store original states in case we need to revert
        const originalLikeActive = likeBtn.classList.contains('active');
        const originalDislikeActive = dislikeBtn.classList.contains('active');
        const originalLikeCount = parseInt(likeBtn.textContent.match(/\d+/)[0]) || 0;
        const originalDislikeCount = parseInt(dislikeBtn.textContent.match(/\d+/)[0]) || 0;

        try {
            // Optimistically update UI
            if (isLike) {
                // Clicking like
                if (originalLikeActive) {
                    // Unlike
                    likeBtn.classList.remove('active');
                    likeBtn.innerHTML = `👍 ${Math.max(0, originalLikeCount - 1)}`;
                } else {
                    // Like (and remove dislike if active)
                    likeBtn.classList.add('active');
                    likeBtn.innerHTML = `👍 ${originalLikeCount + 1}`;
                    if (originalDislikeActive) {
                        dislikeBtn.classList.remove('active');
                        dislikeBtn.innerHTML = `👎 ${Math.max(0, originalDislikeCount - 1)}`;
                    }
                }
            } else {
                // Clicking dislike
                if (originalDislikeActive) {
                    // Un-dislike
                    dislikeBtn.classList.remove('active');
                    dislikeBtn.innerHTML = `👎 ${Math.max(0, originalDislikeCount - 1)}`;
                } else {
                    // Dislike (and remove like if active)
                    dislikeBtn.classList.add('active');
                    dislikeBtn.innerHTML = `👎 ${originalDislikeCount + 1}`;
                    if (originalLikeActive) {
                        likeBtn.classList.remove('active');
                        likeBtn.innerHTML = `👍 ${Math.max(0, originalLikeCount - 1)}`;
                    }
                }
            }

            console.log('🔄 Liking comment:', commentId, 'isLike:', isLike);

            const response = await window.api.likeComment(parseInt(commentId), isLike);
            console.log('📡 Like response:', response);

            if (response.success && response.data) {
                // Update with actual counts from server
                likeBtn.innerHTML = `👍 ${response.data.likeCount || 0}`;
                dislikeBtn.innerHTML = `👎 ${response.data.dislikeCount || 0}`;
            } else if (!response.success) {
                throw new Error(response.message || 'Failed to update like status');
            }
        } catch (error) {
            console.error('Failed to like/dislike comment:', error);

            // Revert UI changes on error
            likeBtn.classList.toggle('active', originalLikeActive);
            dislikeBtn.classList.toggle('active', originalDislikeActive);
            likeBtn.innerHTML = `👍 ${originalLikeCount}`;
            dislikeBtn.innerHTML = `👎 ${originalDislikeCount}`;

            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to update like status');
            }
        }
    },

    handleCommentEdit(event) {
        event.preventDefault();
        event.stopPropagation();

        const button = event.currentTarget;
        const commentId = button.dataset.commentId;
        const commentItem = button.closest('.comment-item');
        const contentDiv = commentItem.querySelector('.comment-content');
        const originalContent = contentDiv.dataset.originalContent;

        // Replace content with textarea
        contentDiv.innerHTML = `
            <div class="edit-comment-form">
                <textarea class="edit-comment-textarea" maxlength="500">${originalContent}</textarea>
                <div class="edit-comment-actions">
                    <button class="btn btn-primary btn-sm save-edit-btn" data-comment-id="${commentId}">Save</button>
                    <button class="btn btn-secondary btn-sm cancel-edit-btn">Cancel</button>
                </div>
            </div>
        `;

        // Bind save and cancel events
        const saveBtn = contentDiv.querySelector('.save-edit-btn');
        const cancelBtn = contentDiv.querySelector('.cancel-edit-btn');
        const textarea = contentDiv.querySelector('.edit-comment-textarea');

        // Auto-resize textarea
        textarea.style.height = 'auto';
        textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px';
        textarea.focus();

        saveBtn.addEventListener('click', () => this.saveCommentEdit(commentId, textarea.value, contentDiv, originalContent));
        cancelBtn.addEventListener('click', () => this.cancelCommentEdit(contentDiv, originalContent));

        // Auto-resize on input
        textarea.addEventListener('input', () => {
            textarea.style.height = 'auto';
            textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px';
        });
    },

    async saveCommentEdit(commentId, newContent, contentDiv, originalContent) {
        if (!newContent.trim()) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Comment cannot be empty');
            }
            return;
        }

        // Check if user is still authenticated before attempting edit
        if (!window.auth || !window.auth.isAuthenticated) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('You must be logged in to edit comments');
            }
            // Restore original content
            contentDiv.innerHTML = window.utils.escapeHtml(originalContent);
            return;
        }

        try {
            console.log('🔄 Updating comment:', commentId, 'content:', newContent.trim());

            const response = await window.api.updateComment(commentId, { content: newContent.trim() });
            console.log('📡 Update response:', response);

            if (response.success) {
                // Update content and original content
                contentDiv.innerHTML = window.utils.escapeHtml(newContent.trim());
                contentDiv.dataset.originalContent = newContent.trim();

                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success('Comment updated successfully');
                }
            } else {
                throw new Error(response.message || 'Failed to update comment');
            }
        } catch (error) {
            console.error('❌ Failed to update comment:', error);

            // Check if this is an authentication error
            if (error.message && (error.message.includes('Authentication required') ||
                                 error.message.includes('Not authenticated') ||
                                 error.message.includes('No active session'))) {

                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.error('Session expired. Please log in again to edit comments.');
                }

                // Clear auth state and redirect to login
                if (window.auth) {
                    window.auth.currentUser = null;
                    window.auth.isAuthenticated = false;
                }

                if (window.forumApp) {
                    window.forumApp.currentUser = null;
                    window.forumApp.isAuthenticated = false;
                    window.forumApp.updateAuthUI();

                    // Redirect to login after a short delay
                    setTimeout(() => {
                        if (window.forumApp.router) {
                            window.forumApp.router.navigate('/login');
                        }
                    }, 2000);
                }
            } else {
                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.error(error.message || 'Failed to update comment');
                }
            }

            // Restore original content on error
            this.cancelCommentEdit(contentDiv, originalContent);
        }
    },

    cancelCommentEdit(contentDiv, originalContent) {
        contentDiv.innerHTML = window.utils.escapeHtml(originalContent);
    },

    async handleCommentDelete(event) {
        event.preventDefault();
        event.stopPropagation();

        const button = event.currentTarget;
        const commentId = button.dataset.commentId;

        if (!confirm('Are you sure you want to delete this comment?')) {
            return;
        }

        try {
            console.log('🔄 Deleting comment:', commentId);

            const response = await window.api.deleteComment(commentId);
            console.log('📡 Delete response:', response);

            if (response.success) {
                // Get post ID before removing comment
                const commentItem = button.closest('.comment-item');
                const commentsContainer = commentItem.closest('.post-comments-container');
                const postId = commentsContainer.id.replace('comments-container-', '');

                // Remove comment from DOM
                commentItem.remove();

                // Update comment count in toggle button
                const toggleBtn = document.querySelector(`[data-post-id="${postId}"].comment-toggle-btn`);
                if (toggleBtn) {
                    const currentCount = parseInt(toggleBtn.textContent.match(/\d+/)[0]) || 0;
                    const newCount = Math.max(0, currentCount - 1);
                    toggleBtn.innerHTML = `💬 ${newCount}`;
                    console.log(`📊 Updated comment count from ${currentCount} to ${newCount} for post ${postId}`);
                } else {
                    console.log(`❌ Could not find toggle button for post ${postId}`);
                }

                // Update comments header if it exists
                const commentsHeader = commentsContainer.querySelector('.comments-header h4');
                if (commentsHeader) {
                    const remainingComments = commentsContainer.querySelectorAll('.comment-item').length;
                    commentsHeader.textContent = `Comments (${remainingComments})`;
                }

                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success('Comment deleted successfully');
                }
            } else {
                throw new Error(response.message || 'Failed to delete comment');
            }
        } catch (error) {
            console.error('Failed to delete comment:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error(error.message || 'Failed to delete comment');
            }
        }
    }
};
