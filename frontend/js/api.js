// api.js - Handles all API requests

const API = {
    // Base URL for API
    baseUrl: '',

    // Helper for making requests
    async request(endpoint, method = 'GET', body = null) {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        try {
            const response = await fetch(endpoint, options);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Something went wrong');
            }

            return data;
        } catch (error) {
            throw error;
        }
    },

    // Auth API
    auth: {
        register: (userData) => API.request('/register', 'POST', userData),
        login: (credentials) => API.request('/login', 'POST', credentials),
        logout: () => API.request('/logout', 'POST'),
        checkSession: async () => {
            // We don't have a dedicated check-session endpoint, 
            // but we can try to get the user info or online users to check auth
            // For now, we'll rely on the cookie and app state, 
            // or we could add a /api/me endpoint.
            // Let's assume we can check if we are logged in by trying to get online users
            // or just rely on the fact that if we get 401 on protected routes, we are logged out.
            // A better way is to add a /api/me endpoint.
            // For this implementation, I'll add a simple check.
            try {
                // Try to get online users as a proxy for session check
                await API.request('/api/online-users');
                return true;
            } catch (e) {
                return false;
            }
        }
    },

    // Posts API
    posts: {
        getAll: () => API.request('/posts'), // Need to update backend to return JSON for this
        create: (postData) => API.request('/posts/create', 'POST', postData),
        getOne: (id) => API.request(`/posts/view?id=${id}`),
    },

    // Comments API
    comments: {
        create: (commentData) => API.request('/comments/create', 'POST', commentData),
    },

    // Messages API
    messages: {
        send: (data) => API.request('/api/messages/send', 'POST', data),
        getHistory: (userId, limit = 50) => API.request(`/api/messages/history?user_id=${userId}&limit=${limit}`),
        getOnlineUsers: () => API.request('/api/online-users'),
    }
};