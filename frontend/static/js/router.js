// Router for SPA navigation

window.Router = class {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        this.isNavigating = false;
        
        // Bind methods
        this.navigate = this.navigate.bind(this);
        this.handlePopState = this.handlePopState.bind(this);
        this.handleLinkClick = this.handleLinkClick.bind(this);
        
        // Set up event listeners
        window.addEventListener('popstate', this.handlePopState);
        document.addEventListener('click', this.handleLinkClick);
    }

    /**
     * Register a route
     */
    addRoute(path, handler, options = {}) {
        this.routes.set(path, {
            handler,
            requiresAuth: options.requiresAuth || false,
            redirectIfAuth: options.redirectIfAuth || false,
            title: options.title || 'Forum'
        });
    }

    /**
     * Navigate to a route
     */
    async navigate(path, pushState = true) {
        if (this.isNavigating) {
            return;
        }

        this.isNavigating = true;
        
        try {
            console.log('🧭 Navigating to:', path);
            
            // Find matching route
            const route = this.findRoute(path);
            if (!route) {
                console.error('❌ Route not found:', path);
                console.log('🔍 Available routes:', Object.keys(this.routes));

                // Show 404 error page
                if (typeof window.showErrorPage === 'function') {
                    console.log('📄 Showing custom 404 error page');
                    window.showErrorPage(404, `The page "${path}" could not be found.`);
                } else {
                    console.log('📄 Falling back to /404 route');
                    this.navigate('/404', false);
                }
                return;
            }

            // Check authentication requirements
            if (route.requiresAuth && !window.auth.isLoggedIn()) {
                console.log('🔒 Route requires authentication, redirecting to login');
                this.navigate('/login', true);
                return;
            }

            if (route.redirectIfAuth && window.auth.isLoggedIn()) {
                console.log('🔓 User already authenticated, redirecting to home');
                this.navigate('/', true);
                return;
            }

            // Update browser history
            if (pushState && window.location.pathname !== path) {
                window.history.pushState({ path }, '', path);
            }

            // Update page title
            document.title = route.title;

            // Update current route
            this.currentRoute = path;

            // Update navigation UI
            this.updateNavigation(path);

            // Update current page in main app
            if (window.forumApp && window.forumApp.setCurrentPage) {
                const pageName = path === '/' ? 'home' : path.substring(1);
                window.forumApp.setCurrentPage(pageName);
            }

            // Execute route handler
            await route.handler(path);

            console.log('✅ Navigation completed:', path);
            
        } catch (error) {
            console.error('❌ Navigation error:', error);
            
            // Show error notification
            if (window.forumApp && window.forumApp.notificationComponent) {
                window.forumApp.notificationComponent.error('Failed to load page');
            }
        } finally {
            this.isNavigating = false;
        }
    }

    /**
     * Find matching route
     */
    findRoute(path) {
        // Exact match first
        if (this.routes.has(path)) {
            return this.routes.get(path);
        }

        // Pattern matching for dynamic routes
        for (const [routePath, route] of this.routes) {
            if (this.matchRoute(routePath, path)) {
                return route;
            }
        }

        return null;
    }

    /**
     * Match route patterns (simple implementation)
     */
    matchRoute(routePath, actualPath) {
        // Convert route pattern to regex
        const pattern = routePath
            .replace(/:\w+/g, '([^/]+)')  // :id -> ([^/]+)
            .replace(/\*/g, '.*');        // * -> .*
        
        const regex = new RegExp(`^${pattern}$`);
        return regex.test(actualPath);
    }

    /**
     * Extract parameters from route
     */
    extractParams(routePath, actualPath) {
        const routeParts = routePath.split('/');
        const actualParts = actualPath.split('/');
        const params = {};

        for (let i = 0; i < routeParts.length; i++) {
            const routePart = routeParts[i];
            if (routePart.startsWith(':')) {
                const paramName = routePart.slice(1);
                params[paramName] = actualParts[i];
            }
        }

        return params;
    }

    /**
     * Handle browser back/forward buttons
     */
    handlePopState(event) {
        const path = event.state?.path || window.location.pathname;
        this.navigate(path, false);
    }

    /**
     * Handle link clicks for SPA navigation
     */
    handleLinkClick(event) {
        const link = event.target.closest('a[data-route]');
        if (!link) return;

        event.preventDefault();
        const path = link.getAttribute('data-route') || link.getAttribute('href');
        this.navigate(path);
    }

    /**
     * Update navigation UI
     */
    updateNavigation(currentPath) {
        // Update active nav links
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            const linkPath = link.getAttribute('data-route') || link.getAttribute('href');
            if (linkPath === currentPath) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
    }

    /**
     * Get current route
     */
    getCurrentRoute() {
        return this.currentRoute;
    }

    /**
     * Initialize router
     */
    init() {
        // Navigate to current path
        const currentPath = window.location.pathname;
        this.navigate(currentPath, false);
    }
};

// Router initialization will be handled by main.js after all components are loaded
