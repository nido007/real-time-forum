// views.js - Handles HTML generation for different views

const Views = {
    // Login View
    getLoginView: () => `
        <div class="auth-container">
            <h2>🔑 Login</h2>
            <form id="login-form">
                <div class="form-group">
                    <label for="login">Username or Email</label>
                    <input type="text" id="login" name="login" required placeholder="Enter username or email">
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required placeholder="Enter password">
                </div>
                <button type="submit" class="btn btn-primary">Login</button>
            </form>
            <p class="auth-link">Don't have an account? <a href="#/register">Register here</a></p>
            <div id="error-message" class="error hidden"></div>
        </div>
    `,

    // Registration View
    getRegisterView: () => `
        <div class="auth-container">
            <h2>📝 Create Account</h2>
            <form id="register-form">
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required minlength="3" maxlength="50">
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required>
                </div>
                <div class="form-row">
                    <div class="form-group half">
                        <label for="first_name">First Name</label>
                        <input type="text" id="first_name" name="first_name" required>
                    </div>
                    <div class="form-group half">
                        <label for="last_name">Last Name</label>
                        <input type="text" id="last_name" name="last_name" required>
                    </div>
                </div>
                <div class="form-row">
                    <div class="form-group half">
                        <label for="age">Age</label>
                        <input type="number" id="age" name="age" required min="1" max="120">
                    </div>
                    <div class="form-group half">
                        <label for="gender">Gender</label>
                        <select id="gender" name="gender" required>
                            <option value="">Select...</option>
                            <option value="Male">Male</option>
                            <option value="Female">Female</option>
                            <option value="Other">Other</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required minlength="6">
                </div>
                <button type="submit" class="btn btn-success">Register</button>
            </form>
            <p class="auth-link">Already have an account? <a href="#/login">Login here</a></p>
            <div id="error-message" class="error hidden"></div>
        </div>
    `,

    // Home/Feed View
    getHomeView: (user) => `
        <div class="dashboard">
            <aside class="sidebar">
                <div class="user-profile-summary">
                    <h3>👋 Hello, ${user.username}</h3>
                </div>
                <div class="sidebar-nav">
                    <a href="#/" class="active">🏠 All Posts</a>
                    <a href="#/my-posts">👤 My Posts</a>
                    <a href="#/liked-posts">❤️ Liked Posts</a>
                </div>
                <div id="online-users-section">
                    <h4>💬 Online Users</h4>
                    <div id="online-users-list">Loading...</div>
                </div>
            </aside>
            <main class="feed">
                <div class="create-post-teaser">
                    <a href="#/create-post" class="btn btn-primary">➕ Create New Post</a>
                </div>
                <div id="posts-feed">
                    <!-- Posts will be loaded here -->
                    <p>Loading posts...</p>
                </div>
            </main>
        </div>
    `,

    // Create Post View
    getCreatePostView: (categories) => {
        let categoriesHTML = '';
        if (categories && categories.length > 0) {
            categoriesHTML = categories.map(cat => `
                <div class="category-item">
                    <input type="checkbox" name="categories" value="${cat.id}" id="cat_${cat.id}">
                    <label for="cat_${cat.id}">${cat.name}</label>
                </div>
            `).join('');
        } else {
            categoriesHTML = '<p>No categories available.</p>';
        }

        return `
        <div class="container">
            <div class="header">
                <h1>➕ Create New Post</h1>
            </div>
            <div class="card form-container">
                <form id="create-post-form">
                    <div class="form-group">
                        <label for="title">📝 Post Title *</label>
                        <input type="text" id="title" name="title" required maxlength="200" 
                               placeholder="Enter an engaging title...">
                    </div>
                    
                    <div class="form-group">
                        <label for="content">📄 Post Content *</label>
                        <textarea id="content" name="content" required 
                                  placeholder="Write your post content here..."></textarea>
                    </div>
                    
                    <div class="form-group">
                        <label>🏷️ Categories *</label>
                        <div class="categories-grid">
                            ${categoriesHTML}
                        </div>
                    </div>
                    
                    <div class="btn-group">
                        <button type="submit" class="btn btn-success">📤 Publish Post</button>
                        <a href="#/" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
                <div id="error-message" class="error hidden"></div>
            </div>
        </div>
        `;
    },

    // Post Detail View
    getPostDetailView: (post, comments, currentUser) => {
        const categoriesHTML = post.categories ? post.categories.map(c => `<span class="badge">${c.name}</span>`).join(' ') : '';
        const date = new Date(post.created_at).toLocaleDateString();

        // Comments HTML
        let commentsHTML = '';
        if (comments && comments.length > 0) {
            commentsHTML = comments.map(comment => `
                <div class="comment-card">
                    <div class="comment-header">
                        <strong>👤 ${comment.author.username}</strong>
                        <span class="comment-meta">• ${new Date(comment.created_at).toLocaleString()}</span>
                    </div>
                    <div class="comment-content">${comment.content}</div>
                </div>
            `).join('');
        } else {
            commentsHTML = '<p class="no-comments">No comments yet. Be the first!</p>';
        }

        return `
        <div class="container">
            <div class="post-detail-card">
                <a href="#/" class="back-link">← Back to Feed</a>
                <h1>${post.title}</h1>
                <div class="post-meta">
                    <span>👤 ${post.author.username}</span> • <span>📅 ${date}</span>
                    <div class="post-categories">${categoriesHTML}</div>
                </div>
                <div class="post-content">
                    ${post.content}
                </div>
                <div class="post-stats">
                    <span>👍 ${post.like_count || 0}</span> • <span>👎 ${post.dislike_count || 0}</span>
                </div>
            </div>

            <div class="comments-section">
                <h3>💬 Comments (${comments ? comments.length : 0})</h3>
                
                <div class="create-comment-form">
                    <form id="create-comment-form" data-post-id="${post.id}">
                        <textarea name="content" required placeholder="Write a comment..."></textarea>
                        <button type="submit" class="btn btn-primary">Post Comment</button>
                    </form>
                </div>

                <div class="comments-list">
                    ${commentsHTML}
                </div>
            </div>
        </div>
        `;
    },

    // Helper to render a single post card
    getPostCard: (post) => {
        const categoriesHTML = post.categories ? post.categories.map(c => `<span class="badge">${c.name}</span>`).join(' ') : '';
        const date = new Date(post.created_at).toLocaleDateString();
        const snippet = post.content.length > 150 ? post.content.substring(0, 150) + '...' : post.content;

        return `
        <div class="card post-card">
            <h3><a href="#/post/${post.id}">${post.title}</a></h3>
            <div class="post-meta">
                <span>👤 ${post.author.username}</span> • <span>📅 ${date}</span>
                <div class="post-categories">${categoriesHTML}</div>
            </div>
            <div class="post-preview">
                <p>${snippet}</p>
            </div>
            <div class="post-stats">
                <span>👍 ${post.like_count || 0}</span> • <span>👎 ${post.dislike_count || 0}</span>
            </div>
        </div>
        `;
    },

    // Chat Window View
    getChatWindow: (userId, username) => `
        <div class="chat-header">
            <div class="chat-user-info">
                <span class="status-indicator online"></span>
                <strong>${username}</strong>
            </div>
            <button class="close-chat">✖</button>
        </div>
        <div id="chat-messages" class="chat-messages">
            <!-- Messages will be loaded here -->
            <p class="loading">Loading history...</p>
        </div>
        <div class="chat-input-area">
            <form id="chat-form">
                <input type="text" placeholder="Type a message..." required autocomplete="off">
                <button type="submit">Send</button>
            </form>
        </div>
    `,

    // Navbar User Info
    getNavbarUserInfo: (user) => {
        if (user) {
            return `
                <span>${user.username}</span>
                <button id="btn-logout" class="btn btn-small btn-secondary">Logout</button>
            `;
        } else {
            return `
                <a href="#/login" class="btn btn-small btn-primary">Login</a>
                <a href="#/register" class="btn btn-small btn-success">Register</a>
            `;
        }
    }
};
