// Individual Post Page Component
window.PostPage = {
    async render(path) {
        window.forumApp.setCurrentPage('post');
        
        // Extract post ID from path
        const postId = path.split('/').pop();
        
        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="post-container">
                <div id="post-content" class="loading-placeholder">
                    Loading post...
                </div>
            </div>
        `;

        await this.loadPost(postId);
    },

    async loadPost(postId) {
        try {
            const response = await window.api.getPost(postId);
            if (response.success) {
                this.renderPost(response.data);
            }
        } catch (error) {
            console.error('Failed to load post:', error);
            const container = document.getElementById('post-content');
            if (container) {
                container.innerHTML = `
                    <div class="error-message">
                        <h2>Post Not Found</h2>
                        <p>The post you're looking for doesn't exist or has been removed.</p>
                        <a href="/posts" data-route="/posts" class="btn btn-primary">Back to Posts</a>
                    </div>
                `;
            }
        }
    },

    renderPost(post) {
        const container = document.getElementById('post-content');
        if (!container) return;

        container.innerHTML = `
            <article class="post-detail">
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
                    <h1 class="post-title">${window.utils.escapeHtml(post.title)}</h1>
                    
                    ${post.categories.length > 0 ? `
                        <div class="post-categories">
                            ${post.categories.map(cat => `<span class="category-tag">${window.utils.escapeHtml(cat)}</span>`).join('')}
                        </div>
                    ` : ''}
                    
                    <div class="post-body">
                        ${window.utils.escapeHtml(post.content).replace(/\n/g, '<br>')}
                    </div>
                    
                    ${post.imagePath ? `
                        <div class="post-image">
                            <img src="${post.imagePath}" alt="Post image" class="post-img">
                        </div>
                    ` : ''}
                </div>
                
                <div class="post-actions">
                    <div class="post-stats">
                        <button class="stat-btn like-btn ${post.userLiked ? 'active' : ''}" 
                                data-post-id="${post.id}" data-action="like">
                            👍 ${post.likeCount}
                        </button>
                        <button class="stat-btn dislike-btn ${post.userDisliked ? 'active' : ''}" 
                                data-post-id="${post.id}" data-action="dislike">
                            👎 ${post.dislikeCount}
                        </button>
                        <span class="stat-item">💬 ${post.commentCount}</span>
                    </div>
                </div>
            </article>
            
            <div class="comments-section">
                <div class="comments-header">
                    <h3>Comments (${post.commentCount})</h3>
                </div>
                
                ${window.auth.isLoggedIn() ? `
                    <form id="comment-form" class="comment-form">
                        <textarea id="comment-content" placeholder="Write a comment..." required></textarea>
                        <button type="submit" class="btn btn-primary">Post Comment</button>
                    </form>
                ` : `
                    <div class="login-prompt">
                        <p><a href="/login" data-route="/login">Login</a> to post a comment</p>
                    </div>
                `}
                
                <div id="comments-list" class="comments-list">
                    <div class="comments-loading">Loading comments...</div>
                </div>
            </div>
        `;

        this.bindEvents(post);

        // Render comments after a short delay for better UX
        setTimeout(() => {
            const commentsContainer = document.getElementById('comments-list');
            if (commentsContainer) {
                commentsContainer.innerHTML = this.renderComments(post.comments || []);
            }
        }, 300);
    },

    renderComments(comments) {
        if (comments.length === 0) {
            return '<div class="no-comments">No comments yet. Be the first to comment!</div>';
        }

        return comments.map(comment => `
            <div class="comment" data-comment-id="${comment.id}">
                <div class="comment-header">
                    <img src="${comment.authorAvatar || '/static/images/default-avatar.png'}"
                         alt="${window.utils.escapeHtml(comment.author)}'s avatar"
                         class="comment-avatar">
                    <div class="comment-info">
                        <span class="comment-author">${window.utils.escapeHtml(comment.author)}</span>
                        <span class="comment-date">${window.utils.formatDate(comment.createdAt)}</span>
                    </div>
                </div>

                <div class="comment-content">
                    ${window.utils.escapeHtml(comment.content).replace(/\n/g, '<br>')}
                </div>

                <div class="comment-actions">
                    ${window.auth.isLoggedIn() ? `
                        <button class="stat-btn like-btn ${comment.userLiked ? 'active' : ''}"
                                data-comment-id="${comment.id}" data-action="like">
                            👍 ${comment.likeCount || 0}
                        </button>
                        <button class="stat-btn dislike-btn ${comment.userDisliked ? 'active' : ''}"
                                data-comment-id="${comment.id}" data-action="dislike">
                            👎 ${comment.dislikeCount || 0}
                        </button>
                    ` : `
                        <span class="stat-item">👍 ${comment.likeCount || 0}</span>
                        <span class="stat-item">👎 ${comment.dislikeCount || 0}</span>
                    `}
                    ${window.auth.isLoggedIn() ? `
                        <button class="reply-btn" data-comment-id="${comment.id}">
                            Reply
                        </button>
                    ` : ''}
                </div>

                ${comment.parentID ? '' : `
                    <div class="comment-replies" id="replies-${comment.id}">
                        <!-- Nested replies would go here -->
                    </div>
                `}
            </div>
        `).join('');
    },

    bindEvents(post) {
        // Like/dislike post
        const postLikeButtons = document.querySelectorAll('[data-post-id] .like-btn, [data-post-id] .dislike-btn');
        postLikeButtons.forEach(button => {
            button.addEventListener('click', this.handlePostLike.bind(this));
        });

        // Like/dislike comments
        const commentLikeButtons = document.querySelectorAll('[data-comment-id] .like-btn, [data-comment-id] .dislike-btn');
        commentLikeButtons.forEach(button => {
            button.addEventListener('click', this.handleCommentLike.bind(this));
        });

        // Comment form
        const commentForm = document.getElementById('comment-form');
        if (commentForm) {
            commentForm.addEventListener('submit', this.handleCommentSubmit.bind(this, post.id));
        }
    },

    async handlePostLike(event) {
        if (!window.auth.isLoggedIn()) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.warning('Please log in to like posts');
            }
            return;
        }

        const button = event.target;
        const postId = parseInt(button.closest('[data-post-id]').dataset.postId);
        const isLike = button.classList.contains('like-btn');

        try {
            const response = await window.api.likePost(postId, isLike);

            if (response.success && response.data) {
                // Update the UI with new counts
                const likeBtn = document.querySelector(`[data-post-id="${postId}"] .like-btn`);
                const dislikeBtn = document.querySelector(`[data-post-id="${postId}"] .dislike-btn`);

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
            console.error('Failed to like post:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to update like status');
            }
        }
    },

    async handleCommentLike(event) {
        if (!window.auth.isLoggedIn()) {
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.warning('Please log in to like comments');
            }
            return;
        }

        const button = event.target;
        const commentId = parseInt(button.closest('[data-comment-id]').dataset.commentId);
        const isLike = button.classList.contains('like-btn');

        try {
            const response = await window.api.likeComment(commentId, isLike);

            if (response.success && response.data) {
                // Update the UI with new counts
                const likeBtn = document.querySelector(`[data-comment-id="${commentId}"] .like-btn`);
                const dislikeBtn = document.querySelector(`[data-comment-id="${commentId}"] .dislike-btn`);

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
            console.error('Failed to like comment:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to update like status');
            }
        }
    },

    async handleCommentSubmit(postId, event) {
        event.preventDefault();
        
        const form = event.target;
        const content = form.querySelector('#comment-content').value.trim();
        
        if (!content) return;

        const submitBtn = form.querySelector('button[type="submit"]');

        try {
            window.utils.setLoading(submitBtn, true, 'Posting...');
            
            await window.api.createComment({
                postId: postId,
                content: content
            });

            // Clear form and reload post
            form.reset();

            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.success('Comment posted successfully');
            }

            // Add the new comment with animation
            await this.loadPost(postId);

            // Scroll to the new comment
            setTimeout(() => {
                const commentsSection = document.querySelector('.comments-section');
                if (commentsSection) {
                    commentsSection.scrollIntoView({
                        behavior: 'smooth',
                        block: 'end'
                    });
                }
            }, 500);

        } catch (error) {
            console.error('Failed to post comment:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to post comment');
            }
        } finally {
            window.utils.setLoading(submitBtn, false);
        }
    }
};
