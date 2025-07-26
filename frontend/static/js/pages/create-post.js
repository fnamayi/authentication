// Create Post Page Component
window.CreatePostPage = {
    async render() {
        window.forumApp.setCurrentPage('create-post');
        
        // Require authentication
        if (!window.auth.requireAuth()) {
            return;
        }

        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = `
            <div class="create-post-container">
                <div class="page-header">
                    <h1>Create New Post</h1>
                    <p>Share your thoughts with the community</p>
                </div>
                
                <form id="create-post-form" class="create-post-form">
                    <div class="form-group">
                        <label for="post-title">Title</label>
                        <input 
                            type="text" 
                            id="post-title" 
                            name="title" 
                            required 
                            placeholder="Enter a descriptive title for your post"
                            maxlength="200"
                        >
                    </div>
                    
                    <div class="form-group">
                        <label for="post-categories">Categories</label>
                        <div class="categories-section">
                            <div class="predefined-categories">
                                <label class="categories-label">Choose Categories:</label>
                                <div class="category-checkboxes">
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="General" name="predefined-category">
                                        <span>General</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Technology" name="predefined-category">
                                        <span>Technology</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Sports" name="predefined-category">
                                        <span>Sports</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Entertainment" name="predefined-category">
                                        <span>Entertainment</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Gaming" name="predefined-category">
                                        <span>Gaming</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Music" name="predefined-category">
                                        <span>Music</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Movies" name="predefined-category">
                                        <span>Movies</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Food" name="predefined-category">
                                        <span>Food</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Travel" name="predefined-category">
                                        <span>Travel</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Health" name="predefined-category">
                                        <span>Health</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Education" name="predefined-category">
                                        <span>Education</span>
                                    </label>
                                    <label class="category-checkbox">
                                        <input type="checkbox" value="Science" name="predefined-category">
                                        <span>Science</span>
                                    </label>
                                </div>
                            </div>
                        </div>
                        <small class="form-help">Select relevant categories for your post</small>
                    </div>
                    
                    <div class="form-group">
                        <label for="post-content">Content</label>
                        <textarea 
                            id="post-content" 
                            name="content" 
                            required 
                            placeholder="Write your post content here..."
                            rows="10"
                        ></textarea>
                    </div>
                    
                    <div class="form-actions">
                        <button type="button" class="btn btn-outline" onclick="history.back()">
                            Cancel
                        </button>
                        <button type="submit" class="btn btn-primary">
                            Create Post
                        </button>
                    </div>
                </form>
            </div>
        `;

        this.bindEvents();
    },

    bindEvents() {
        const form = document.getElementById('create-post-form');
        if (form) {
            form.addEventListener('submit', this.handleSubmit.bind(this));
        }

        // Character counter for title
        const titleInput = document.getElementById('post-title');
        if (titleInput) {
            titleInput.addEventListener('input', this.updateCharacterCount.bind(this));
        }

        // Simple category counter
        this.updateCategoryCounter();
        const categoryCheckboxes = document.querySelectorAll('input[name="predefined-category"]');
        categoryCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', this.updateCategoryCounter.bind(this));
        });
    },

    updateCharacterCount(event) {
        const input = event.target;
        const maxLength = input.maxLength;
        const currentLength = input.value.length;

        // Add character counter if it doesn't exist
        let counter = input.parentNode.querySelector('.char-counter');
        if (!counter) {
            counter = document.createElement('small');
            counter.className = 'char-counter form-help';
            input.parentNode.appendChild(counter);
        }

        counter.textContent = `${currentLength}/${maxLength} characters`;
        counter.style.color = currentLength > maxLength * 0.9 ? 'var(--warning-color)' : 'var(--text-muted)';
    },



    updateCategoryCounter() {
        const selectedCategories = document.querySelectorAll('input[name="predefined-category"]:checked');
        const totalSelected = selectedCategories.length;

        // Add or update category counter
        let counter = document.querySelector('.category-counter');
        if (!counter) {
            counter = document.createElement('div');
            counter.className = 'category-counter';
            const categoriesSection = document.querySelector('.categories-section');
            categoriesSection.appendChild(counter);
        }

        counter.innerHTML = `
            <span class="counter-text">
                ${totalSelected} ${totalSelected === 1 ? 'category' : 'categories'} selected
            </span>
            ${totalSelected > 0 ? '<button type="button" class="clear-all-btn">Clear All</button>' : ''}
        `;

        // Bind clear all button
        const clearAllBtn = counter.querySelector('.clear-all-btn');
        if (clearAllBtn) {
            clearAllBtn.addEventListener('click', this.clearAllCategories.bind(this));
        }
    },

    clearAllCategories() {
        // Uncheck all categories
        const checkboxes = document.querySelectorAll('input[name="predefined-category"]');
        checkboxes.forEach(checkbox => {
            checkbox.checked = false;
        });

        this.updateCategoryCounter();
    },

    async handleSubmit(event) {
        event.preventDefault();

        const form = event.target;
        const formData = new FormData(form);

        const title = formData.get('title').trim();
        const content = formData.get('content').trim();

        // Collect selected categories
        const categories = [];
        const checkboxes = form.querySelectorAll('input[name="predefined-category"]:checked');
        checkboxes.forEach(checkbox => {
            categories.push(checkbox.value);
        });

        const submitBtn = form.querySelector('button[type="submit"]');

        try {
            window.utils.setLoading(submitBtn, true, 'Creating Post...');
            
            const response = await window.api.createPost({
                title,
                content,
                categories
            });

            if (response.success) {
                if (window.forumApp.notificationComponent) {
                    window.forumApp.notificationComponent.success('Post created successfully!');
                }
                
                // Navigate to the new post
                if (window.forumApp.router) {
                    window.forumApp.router.navigate(`/post/${response.data.id}`);
                }
            }

        } catch (error) {
            console.error('Failed to create post:', error);
            if (window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error(error.message || 'Failed to create post');
            }
        } finally {
            window.utils.setLoading(submitBtn, false);
        }
    }
};
