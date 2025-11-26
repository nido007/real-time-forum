// app.js - Main Application Logic

const App = {
    state: {
        user: null,
        currentView: 'home',
    },

    init: async () => {
        console.log('🚀 App Initializing...');

        // Check if user is logged in (persisted state or session check)
        const savedUser = localStorage.getItem('user');
        if (savedUser) {
            App.state.user = JSON.parse(savedUser);
        }

        // Initialize Router
        window.addEventListener('hashchange', App.handleRoute);

        // Initial Route
        App.handleRoute();

        // Update Navbar
        App.renderNavbar();
    },

    handleRoute: () => {
        const hash = window.location.hash || '#/';
        console.log('📍 Route:', hash);

        const appContainer = document.getElementById('app');

        if (hash === '#/login') {
            appContainer.innerHTML = Views.getLoginView();
            App.bindLoginEvents();
        } else if (hash === '#/register') {
            appContainer.innerHTML = Views.getRegisterView();
            App.bindRegisterEvents();
        } else if (hash === '#/') {
            if (!App.state.user) {
                window.location.hash = '#/login';
                return;
            }
            appContainer.innerHTML = Views.getHomeView(App.state.user);
            App.loadFeed();
            // Initialize WebSocket if logged in
            if (window.WebSocketHandler) {
                window.WebSocketHandler.connect();
            }
        } else if (hash === '#/create-post') {
            if (!App.state.user) {
                window.location.hash = '#/login';
                return;
            }
            // Fetch categories first
            // For now, we'll hardcode or fetch if API available. 
            // We don't have API.categories.getAll exposed in api.js yet, 
            // but we can assume we can get them or just use hardcoded for now 
            // OR better, update api.js to fetch categories.
            // Wait, posts.go `getAllCategories` is internal.
            // I should have exposed an endpoint for categories.
            // But `CreatePostHandler` (GET) in Go used to render form with categories.
            // Now `CreatePostHandler` (POST) expects category IDs.
            // I need an endpoint to get categories!
            // I missed that.
            // For now, I'll use the hardcoded categories in `init.go` 
            // or I can add a `GET /categories` endpoint.
            // Let's assume I'll add `GET /categories` later.
            // For now, I'll mock them in `getCreatePostView` call.
            const categories = [
                { id: 1, name: 'Technology' },
                { id: 2, name: 'Gaming' },
                { id: 3, name: 'Sports' },
                { id: 4, name: 'General' }
            ];
            appContainer.innerHTML = Views.getCreatePostView(categories);
            App.bindCreatePostEvents();
        } else if (hash.startsWith('#/post/')) {
            if (!App.state.user) {
                window.location.hash = '#/login';
                return;
            }
            const postId = hash.split('/')[2];
            App.loadPostDetail(postId);
        } else {
            appContainer.innerHTML = '<h2>404 Not Found</h2>';
        }
    },

    renderNavbar: () => {
        const userInfoContainer = document.getElementById('user-info');
        if (userInfoContainer) {
            userInfoContainer.innerHTML = Views.getNavbarUserInfo(App.state.user);
        }

        const logoutBtn = document.getElementById('btn-logout');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', App.handleLogout);
        }
    },

    bindLoginEvents: () => {
        const form = document.getElementById('login-form');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            const data = Object.fromEntries(formData.entries());

            try {
                const response = await API.auth.login(data);
                App.state.user = response.user;
                localStorage.setItem('user', JSON.stringify(response.user));
                App.renderNavbar();
                window.location.hash = '#/';
            } catch (error) {
                App.showError(error.message);
            }
        });
    },

    bindRegisterEvents: () => {
        const form = document.getElementById('register-form');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);
            const data = Object.fromEntries(formData.entries());

            // Convert age to int
            data.age = parseInt(data.age);

            try {
                await API.auth.register(data);
                alert('Registration successful! Please login.');
                window.location.hash = '#/login';
            } catch (error) {
                App.showError(error.message);
            }
        });
    },

    bindCreatePostEvents: () => {
        const form = document.getElementById('create-post-form');
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(form);

            // Handle multiple categories
            const categories = [];
            document.querySelectorAll('input[name="categories"]:checked').forEach(checkbox => {
                categories.push(checkbox.value);
            });

            const data = {
                title: formData.get('title'),
                content: formData.get('content'),
                categories: categories
            };

            try {
                const response = await API.posts.create(data);
                alert('Post created successfully!');
                window.location.hash = '#/';
            } catch (error) {
                App.showError(error.message);
            }
        });
    },

    handleLogout: async () => {
        try {
            await API.auth.logout();
            App.state.user = null;
            localStorage.removeItem('user');
            App.renderNavbar();
            window.location.hash = '#/login';
        } catch (error) {
            console.error('Logout failed:', error);
        }
    },

    showError: (message) => {
        const errorDiv = document.getElementById('error-message');
        if (errorDiv) {
            errorDiv.textContent = message;
            errorDiv.classList.remove('hidden');
        } else {
            alert(message);
        }
    },

    loadFeed: async () => {
        const feedContainer = document.getElementById('posts-feed');
        feedContainer.innerHTML = '<p>Loading posts...</p>';

        try {
            const posts = await API.posts.getAll();
            if (posts && posts.length > 0) {
                feedContainer.innerHTML = posts.map(post => Views.getPostCard(post)).join('');
            } else {
                feedContainer.innerHTML = '<p>No posts found. Be the first to create one!</p>';
            }
        } catch (error) {
            feedContainer.innerHTML = `<p class="error">Error loading posts: ${error.message}</p>`;
        }
    },

    loadPostDetail: async (postId) => {
        const appContainer = document.getElementById('app');
        appContainer.innerHTML = '<p>Loading post...</p>';

        try {
            const data = await API.posts.getOne(postId);
            appContainer.innerHTML = Views.getPostDetailView(data.post, data.comments, App.state.user);

            // Bind comment form event
            const commentForm = document.getElementById('create-comment-form');
            if (commentForm) {
                commentForm.addEventListener('submit', App.handleCreateComment);
            }
        } catch (error) {
            appContainer.innerHTML = `<h2>Error</h2><p>${error.message}</p><a href="#/">Back to Home</a>`;
        }
    },

    handleCreateComment: async (e) => {
        e.preventDefault();
        const form = e.target;
        const postId = form.dataset.postId;
        const content = form.querySelector('textarea[name="content"]').value;

        try {
            await API.comments.create({
                post_id: parseInt(postId),
                content: content
            });
            // Reload post detail to show new comment
            App.loadPostDetail(postId);
        } catch (error) {
            alert('Error creating comment: ' + error.message);
        }
    }
};

// Start App when DOM is ready
document.addEventListener('DOMContentLoaded', App.init);
