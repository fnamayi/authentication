// Utility functions for the Forum application

window.utils = {
    /**
     * Escape HTML to prevent XSS attacks
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    /**
     * Format date for display
     */
    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) {
            return 'Just now';
        } else if (diffMins < 60) {
            return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
        } else if (diffHours < 24) {
            return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
        } else if (diffDays < 7) {
            return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
        } else {
            return date.toLocaleDateString();
        }
    },

    /**
     * Format time for messages (as specified in prompt: [YYYY-MM-DD HH:MM])
     */
    formatTime(dateString) {
        const date = new Date(dateString);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        
        return `[${year}-${month}-${day} ${hours}:${minutes}]`;
    },

    /**
     * Debounce function to limit API calls
     */
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },

    /**
     * Throttle function to limit function calls
     */
    throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    },

    /**
     * Generate a random ID
     */
    generateId() {
        return Math.random().toString(36).substr(2, 9);
    },

    /**
     * Validate email format
     */
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    },

    /**
     * Validate password strength
     */
    isValidPassword(password) {
        return password && password.length >= 8;
    },

    /**
     * Sanitize input for display
     */
    sanitizeInput(input) {
        return input.trim().replace(/\s+/g, ' ');
    },

    /**
     * Format file size
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    },

    /**
     * Copy text to clipboard
     */
    async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = text;
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();
            try {
                document.execCommand('copy');
                document.body.removeChild(textArea);
                return true;
            } catch (err) {
                document.body.removeChild(textArea);
                return false;
            }
        }
    },

    /**
     * Show/hide loading state on elements
     */
    setLoading(element, isLoading, loadingText = 'Loading...') {
        if (!element) return;
        
        if (isLoading) {
            element.dataset.originalText = element.textContent;
            element.textContent = loadingText;
            element.disabled = true;
            element.classList.add('loading');
        } else {
            element.textContent = element.dataset.originalText || element.textContent;
            element.disabled = false;
            element.classList.remove('loading');
            delete element.dataset.originalText;
        }
    },

    /**
     * Smooth scroll to element
     */
    scrollToElement(element, offset = 0) {
        if (!element) return;
        
        const elementPosition = element.getBoundingClientRect().top;
        const offsetPosition = elementPosition + window.pageYOffset - offset;

        window.scrollTo({
            top: offsetPosition,
            behavior: 'smooth'
        });
    },

    /**
     * Check if element is in viewport
     */
    isInViewport(element) {
        const rect = element.getBoundingClientRect();
        return (
            rect.top >= 0 &&
            rect.left >= 0 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth)
        );
    },

    /**
     * Local storage helpers
     */
    storage: {
        set(key, value) {
            try {
                localStorage.setItem(key, JSON.stringify(value));
                return true;
            } catch (e) {
                console.error('Failed to save to localStorage:', e);
                return false;
            }
        },

        get(key, defaultValue = null) {
            try {
                const item = localStorage.getItem(key);
                return item ? JSON.parse(item) : defaultValue;
            } catch (e) {
                console.error('Failed to read from localStorage:', e);
                return defaultValue;
            }
        },

        remove(key) {
            try {
                localStorage.removeItem(key);
                return true;
            } catch (e) {
                console.error('Failed to remove from localStorage:', e);
                return false;
            }
        },

        clear() {
            try {
                localStorage.clear();
                return true;
            } catch (e) {
                console.error('Failed to clear localStorage:', e);
                return false;
            }
        }
    },

    /**
     * URL helpers
     */
    url: {
        getParams() {
            return new URLSearchParams(window.location.search);
        },

        getParam(name, defaultValue = null) {
            const params = this.getParams();
            return params.get(name) || defaultValue;
        },

        setParam(name, value) {
            const url = new URL(window.location);
            url.searchParams.set(name, value);
            window.history.replaceState({}, '', url);
        },

        removeParam(name) {
            const url = new URL(window.location);
            url.searchParams.delete(name);
            window.history.replaceState({}, '', url);
        }
    },

    /**
     * DOM helpers
     */
    dom: {
        createElement(tag, className = '', textContent = '') {
            const element = document.createElement(tag);
            if (className) element.className = className;
            if (textContent) element.textContent = textContent;
            return element;
        },

        removeElement(element) {
            if (element && element.parentNode) {
                element.parentNode.removeChild(element);
            }
        },

        toggleClass(element, className, force = null) {
            if (!element) return;
            if (force !== null) {
                element.classList.toggle(className, force);
            } else {
                element.classList.toggle(className);
            }
        },

        hasClass(element, className) {
            return element && element.classList.contains(className);
        }
    },

    /**
     * Event helpers
     */
    events: {
        on(element, event, handler, options = {}) {
            if (element) {
                element.addEventListener(event, handler, options);
            }
        },

        off(element, event, handler) {
            if (element) {
                element.removeEventListener(event, handler);
            }
        },

        once(element, event, handler) {
            if (element) {
                element.addEventListener(event, handler, { once: true });
            }
        },

        trigger(element, event, data = {}) {
            if (element) {
                const customEvent = new CustomEvent(event, { detail: data });
                element.dispatchEvent(customEvent);
            }
        }
    }
};
