package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"real-time-forum/internal/database"
	"real-time-forum/internal/middleware"
)

// PostsHandler handles all post-related HTTP requests
type PostsHandler struct {
	db             *sql.DB
	authMiddleware *middleware.AuthMiddleware
}

// NewPostsHandler creates a new posts handler
func NewPostsHandler(db *sql.DB, authMiddleware *middleware.AuthMiddleware) *PostsHandler {
	return &PostsHandler{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

// ListPostsHandler displays all posts with filtering options
func (h *PostsHandler) ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := h.authMiddleware.GetCurrentUser(r)

	// Get filter parameters
	categoryID := r.URL.Query().Get("category")
	filter := r.URL.Query().Get("filter") // "my-posts", "liked-posts"

	// Get posts based on filters
	posts, err := h.getPosts(categoryID, filter, currentUser)
	if err != nil {
		h.showError(w, "Error loading posts: "+err.Error())
		return
	}

	// Get all categories for filter dropdown
	categories, err := h.getAllCategories()
	if err != nil {
		h.showError(w, "Error loading categories: "+err.Error())
		return
	}

	// Render posts page
	h.renderPostsList(w, posts, categories, currentUser, categoryID, filter)
}

// CreatePostHandler handles post creation (GET: form, POST: process)
func (h *PostsHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.showCreatePostForm(w, currentUser)
	case http.MethodPost:
		h.processCreatePost(w, r, currentUser)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ViewPostHandler displays a single post with comments
func (h *PostsHandler) ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		h.showError(w, "Post ID is required")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.showError(w, "Invalid post ID")
		return
	}

	currentUser := h.authMiddleware.GetCurrentUser(r)

	// Get post details
	post, err := h.getPostByID(postID, currentUser)
	if err != nil {
		if err == sql.ErrNoRows {
			h.showError(w, "Post not found")
		} else {
			h.showError(w, "Error loading post: "+err.Error())
		}
		return
	}

	// Get comments for this post
	comments, err := h.getCommentsByPostID(postID, currentUser)
	if err != nil {
		h.showError(w, "Error loading comments: "+err.Error())
		return
	}

	// Render post view
	h.renderPostView(w, post, comments, currentUser)
}

// showCreatePostForm displays the post creation form
func (h *PostsHandler) showCreatePostForm(w http.ResponseWriter, user *database.User) {
	categories, err := h.getAllCategories()
	if err != nil {
		h.showError(w, "Error loading categories: "+err.Error())
		return
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Create New Post - Forum</title>
		<style>
			%s
			.form-container { max-width: 800px; margin: 0 auto; }
			.form-group { margin-bottom: 20px; }
			.form-group label { display: block; margin-bottom: 8px; font-weight: 600; }
			.form-group input, .form-group textarea, .form-group select { 
				width: 100%%; padding: 12px; border: 2px solid #ddd; border-radius: 6px; 
				font-size: 16px; box-sizing: border-box;
			}
			.form-group textarea { min-height: 200px; resize: vertical; font-family: inherit; }
			.categories-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; }
			.category-item { display: flex; align-items: center; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
			.category-item input[type="checkbox"] { margin-right: 8px; }
			.btn-group { display: flex; gap: 10px; flex-wrap: wrap; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>➕ Create New Post</h1>
				<p>Share your thoughts with the community, %s!</p>
			</div>
			
			<div class="card form-container">
				<form method="POST" action="/posts/create">
					<div class="form-group">
						<label for="title">📝 Post Title *</label>
						<input type="text" id="title" name="title" required maxlength="200" 
							   placeholder="Enter an engaging title for your post...">
						<small>Maximum 200 characters</small>
					</div>
					
					<div class="form-group">
						<label for="content">📄 Post Content *</label>
						<textarea id="content" name="content" required 
								  placeholder="Write your post content here... Share your thoughts, ask questions, or start a discussion!"></textarea>
						<small>Minimum 10 characters required</small>
					</div>
					
					<div class="form-group">
						<label>🏷️ Categories * (select one or more)</label>
						<div class="categories-grid">
	`, getCommonCSS(), template.HTMLEscapeString(user.Username))

	for _, category := range categories {
		html += fmt.Sprintf(`
							<div class="category-item">
								<input type="checkbox" name="categories" value="%d" id="cat_%d">
								<label for="cat_%d"><strong>%s</strong><br><small>%s</small></label>
							</div>
		`, category.ID, category.ID, category.ID,
			template.HTMLEscapeString(category.Name),
			template.HTMLEscapeString(category.Description))
	}

	html += `
						</div>
						<small>Choose categories that best describe your post topic</small>
					</div>
					
					<div class="btn-group">
						<button type="submit" class="btn btn-success">📤 Publish Post</button>
						<a href="/posts" class="btn btn-secondary">Cancel</a>
						<a href="/" class="btn btn-primary">🏠 Home</a>
					</div>
				</form>
			</div>
		</div>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// processCreatePost handles post creation form submission
func (h *PostsHandler) processCreatePost(w http.ResponseWriter, r *http.Request, user *database.User) {
	err := r.ParseForm()
	if err != nil {
		h.showError(w, "Invalid form data")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryIDs := r.Form["categories"]

	// Validate input
	if title == "" || content == "" {
		h.showError(w, "Title and content are required")
		return
	}

	if len(title) > 200 {
		h.showError(w, "Title must be 200 characters or less")
		return
	}

	if len(content) < 10 {
		h.showError(w, "Content must be at least 10 characters long")
		return
	}

	if len(categoryIDs) == 0 {
		h.showError(w, "Please select at least one category")
		return
	}

	// Create post
	postID, err := h.createPost(user.ID, title, content, categoryIDs)
	if err != nil {
		h.showError(w, "Error creating post: "+err.Error())
		return
	}

	// Show success and redirect
	h.showCreatePostSuccess(w, postID, title)
}

// getPosts retrieves posts based on filters
func (h *PostsHandler) getPosts(categoryID, filter string, currentUser *database.User) ([]database.Post, error) {
	var posts []database.Post
	var query string
	var args []interface{}

	baseQuery := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`

	var conditions []string

	// Apply category filter
	if categoryID != "" {
		conditions = append(conditions, "p.id IN (SELECT post_id FROM post_categories WHERE category_id = ?)")
		args = append(args, categoryID)
	}

	// Apply user-specific filters
	if currentUser != nil && filter != "" {
		switch filter {
		case "my-posts":
			conditions = append(conditions, "p.user_id = ?")
			args = append(args, currentUser.ID)
		case "liked-posts":
			conditions = append(conditions, "p.id IN (SELECT post_id FROM likes WHERE user_id = ? AND post_id IS NOT NULL AND is_like = 1)")
			args = append(args, currentUser.ID)
		}
	}

	// Build final query
	if len(conditions) > 0 {
		query = baseQuery + " WHERE " + strings.Join(conditions, " AND ")
	} else {
		query = baseQuery
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post database.Post
		var authorUsername string

		// FIXED: Scan into separate variables first
		err := rows.Scan(&post.ID, &post.UserID, &authorUsername, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		// FIXED: Initialize Author struct properly AFTER scanning
		post.Author = &database.User{
			ID:       post.UserID,
			Username: authorUsername,
		}

		// Get categories for this post
		post.Categories, err = h.getCategoriesByPostID(post.ID)
		if err != nil {
			return nil, err
		}

		// Get vote counts
		post.LikeCount, post.DislikeCount, post.UserVote = h.getVoteStats("post", post.ID, currentUser)

		posts = append(posts, post)
	}

	return posts, nil
}

// getPostByID retrieves a single post by ID
func (h *PostsHandler) getPostByID(postID int, currentUser *database.User) (*database.Post, error) {
	var post database.Post
	post.Author = &database.User{}

	err := h.db.QueryRow(`
		SELECT p.id, p.user_id, u.username, u.email, p.title, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, postID).Scan(&post.ID, &post.UserID, &post.Author.Username, &post.Author.Email,
		&post.Title, &post.Content, &post.CreatedAt)

	if err != nil {
		return nil, err
	}

	post.Author.ID = post.UserID

	// Get categories
	post.Categories, err = h.getCategoriesByPostID(post.ID)
	if err != nil {
		return nil, err
	}

	// Get vote stats
	post.LikeCount, post.DislikeCount, post.UserVote = h.getVoteStats("post", post.ID, currentUser)

	return &post, nil
}

// createPost creates a new post in the database
func (h *PostsHandler) createPost(userID int, title, content string, categoryIDs []string) (int64, error) {
	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert post
	result, err := tx.Exec(`
		INSERT INTO posts (user_id, title, content) 
		VALUES (?, ?, ?)
	`, userID, title, content)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert post-category relationships
	for _, categoryIDStr := range categoryIDs {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			continue // Skip invalid category IDs
		}

		_, err = tx.Exec(`
			INSERT INTO post_categories (post_id, category_id) 
			VALUES (?, ?)
		`, postID, categoryID)
		if err != nil {
			return 0, err
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return postID, nil
}

// getAllCategories retrieves all available categories
func (h *PostsHandler) getAllCategories() ([]database.Category, error) {
	rows, err := h.db.Query(`
		SELECT id, name, description, created_at 
		FROM categories 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []database.Category
	for rows.Next() {
		var category database.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// getCategoriesByPostID retrieves categories for a specific post
func (h *PostsHandler) getCategoriesByPostID(postID int) ([]database.Category, error) {
	rows, err := h.db.Query(`
		SELECT c.id, c.name, c.description, c.created_at
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
		ORDER BY c.name
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []database.Category
	for rows.Next() {
		var category database.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// getVoteStats retrieves vote counts and user vote status
func (h *PostsHandler) getVoteStats(targetType string, targetID int, currentUser *database.User) (int, int, *bool) {
	var likeCount, dislikeCount int
	var userVote *bool

	// Get vote counts based on target type
	var countQuery string
	if targetType == "post" {
		countQuery = `
			SELECT 
				COUNT(CASE WHEN is_like = 1 THEN 1 END) as likes,
				COUNT(CASE WHEN is_like = 0 THEN 1 END) as dislikes
			FROM likes 
			WHERE post_id = ?
		`
	} else if targetType == "comment" {
		countQuery = `
			SELECT 
				COUNT(CASE WHEN is_like = 1 THEN 1 END) as likes,
				COUNT(CASE WHEN is_like = 0 THEN 1 END) as dislikes
			FROM likes 
			WHERE comment_id = ?
		`
	}

	err := h.db.QueryRow(countQuery, targetID).Scan(&likeCount, &dislikeCount)
	if err != nil {
		return 0, 0, nil
	}

	// Get current user's vote if logged in
	if currentUser != nil {
		var isLike sql.NullBool
		var userVoteQuery string

		if targetType == "post" {
			userVoteQuery = `SELECT is_like FROM likes WHERE user_id = ? AND post_id = ?`
		} else if targetType == "comment" {
			userVoteQuery = `SELECT is_like FROM likes WHERE user_id = ? AND comment_id = ?`
		}

		err = h.db.QueryRow(userVoteQuery, currentUser.ID, targetID).Scan(&isLike)
		if err == nil && isLike.Valid {
			userVote = &isLike.Bool
		}
	}

	return likeCount, dislikeCount, userVote
}

// getCommentsByPostID retrieves comments for a specific post
func (h *PostsHandler) getCommentsByPostID(postID int, currentUser *database.User) ([]database.Comment, error) {
	rows, err := h.db.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []database.Comment
	for rows.Next() {
		var comment database.Comment
		comment.Author = &database.User{}

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID,
			&comment.Author.Username, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}

		comment.Author.ID = comment.UserID

		// Get vote stats for this comment
		comment.LikeCount, comment.DislikeCount, comment.UserVote = h.getVoteStats("comment", comment.ID, currentUser)

		comments = append(comments, comment)
	}

	return comments, nil
}

// Helper function to get common CSS styles
func getCommonCSS() string {
	return `
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { 
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			line-height: 1.6; color: #333; background-color: #f8f9fa;
		}
		.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
		.header { text-align: center; margin-bottom: 40px; }
		.header h1 { color: #2c3e50; margin-bottom: 10px; }
		.header p { color: #7f8c8d; font-size: 1.1em; }
		.card { 
			background: white; padding: 30px; margin: 20px 0; 
			border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);
		}
		.btn { 
			display: inline-block; padding: 12px 24px; margin: 8px 8px 8px 0;
			text-decoration: none; border-radius: 6px; font-weight: 600;
			transition: transform 0.2s, box-shadow 0.2s; border: none; cursor: pointer;
		}
		.btn:hover { transform: translateY(-2px); box-shadow: 0 4px 15px rgba(0,0,0,0.2); }
		.btn-primary { background: #3498db; color: white; }
		.btn-success { background: #2ecc71; color: white; }
		.btn-secondary { background: #95a5a6; color: white; }
		.btn-danger { background: #e74c3c; color: white; }
		.error-container { background: #f8d7da; padding: 20px; border-radius: 8px; 
						   border: 1px solid #f5c6cb; margin: 20px 0; }
		small { color: #6c757d; font-size: 0.9em; }
	`
}

// Render functions for different views
func (h *PostsHandler) renderPostsList(w http.ResponseWriter, posts []database.Post, categories []database.Category, currentUser *database.User, selectedCategory, selectedFilter string) {
	userSection := ""
	if currentUser != nil {
		userSection = fmt.Sprintf(`
		<div class="user-actions">
			<p>Welcome, <strong>%s</strong>!</p>
			<a href="/posts/create" class="btn btn-success">➕ Create New Post</a>
			<a href="/logout" class="btn btn-secondary">🚪 Logout</a>
		</div>`, template.HTMLEscapeString(currentUser.Username))
	} else {
		userSection = `
		<div class="guest-actions">
			<p><a href="/login">Login</a> to create posts and participate in discussions!</p>
		</div>`
	}

	// Generate filter options
	filterOptions := ""
	if currentUser != nil {
		filterOptions = fmt.Sprintf(`
		<div class="filters">
			<a href="/posts" class="btn %s">All Posts</a>
			<a href="/posts?filter=my-posts" class="btn %s">My Posts</a>
			<a href="/posts?filter=liked-posts" class="btn %s">Liked Posts</a>
		</div>`,
			getActiveClass(selectedFilter == ""),
			getActiveClass(selectedFilter == "my-posts"),
			getActiveClass(selectedFilter == "liked-posts"))
	}

	// Generate posts HTML
	postsHTML := ""
	for _, post := range posts {
		categoriesStr := ""
		for i, cat := range post.Categories {
			if i > 0 {
				categoriesStr += ", "
			}
			categoriesStr += cat.Name
		}

		userVoteIcon := ""
		if post.UserVote != nil {
			if *post.UserVote {
				userVoteIcon = "👍"
			} else {
				userVoteIcon = "👎"
			}
		}

		postsHTML += fmt.Sprintf(`
		<div class="card post-card">
			<h3><a href="/posts/view?id=%d">%s</a></h3>
			<div class="post-meta">
				<p>👤 by <strong>%s</strong> • 📅 %s • 🏷️ %s</p>
			</div>
			<div class="post-preview">
				<p>%s</p>
			</div>
			<div class="post-stats">
				<span>👍 %d</span> • <span>👎 %d</span> %s
			</div>
		</div>`,
			post.ID, template.HTMLEscapeString(post.Title),
			template.HTMLEscapeString(post.Author.Username),
			post.CreatedAt.Format("Jan 2, 2006"),
			template.HTMLEscapeString(categoriesStr),
			template.HTMLEscapeString(truncateText(post.Content, 200)),
			post.LikeCount, post.DislikeCount, userVoteIcon)
	}

	if len(posts) == 0 {
		postsHTML = `
		<div class="card" style="text-align: center; padding: 60px;">
			<h3>📭 No posts found</h3>
			<p>Be the first to create a post or try a different filter!</p>
		</div>`
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Forum Posts</title>
		<style>
			%s
			.post-card { margin-bottom: 20px; }
			.post-card h3 a { color: #2c3e50; text-decoration: none; }
			.post-card h3 a:hover { color: #3498db; }
			.post-meta { color: #7f8c8d; margin: 10px 0; }
			.post-preview { margin: 15px 0; }
			.post-stats { color: #7f8c8d; font-size: 0.9em; }
			.filters { margin: 20px 0; }
			.btn.active { background: #2c3e50; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>📝 Forum Posts</h1>
				<p>Discover and share interesting discussions</p>
			</div>
			
			<div class="card">
				%s
				%s
			</div>
			
			%s
			
			<div>
				<a href="/" class="btn btn-secondary">🏠 Back to Home</a>
			</div>
		</div>
	</body>
	</html>`, getCommonCSS(), userSection, filterOptions, postsHTML)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func (h *PostsHandler) renderPostView(w http.ResponseWriter, post *database.Post, comments []database.Comment, currentUser *database.User) {
	// Generate vote buttons
	voteButtons := ""
	if currentUser != nil {
		likeActive := ""
		dislikeActive := ""
		if post.UserVote != nil {
			if *post.UserVote {
				likeActive = "active"
			} else {
				dislikeActive = "active"
			}
		}

		voteButtons = fmt.Sprintf(`
		<div class="vote-buttons">
			<form method="POST" action="/vote" style="display: inline;">
				<input type="hidden" name="type" value="like">
				<input type="hidden" name="target" value="post">
				<input type="hidden" name="target_id" value="%d">
				<input type="hidden" name="redirect" value="/posts/view?id=%d">
				<button type="submit" class="btn vote-btn like %s">👍 Like</button>
			</form>
			<form method="POST" action="/vote" style="display: inline;">
				<input type="hidden" name="type" value="dislike">
				<input type="hidden" name="target" value="post">
				<input type="hidden" name="target_id" value="%d">
				<input type="hidden" name="redirect" value="/posts/view?id=%d">
				<button type="submit" class="btn vote-btn dislike %s">👎 Dislike</button>
			</form>
		</div>`, post.ID, post.ID, likeActive, post.ID, post.ID, dislikeActive)
	} else {
		voteButtons = `<p><a href="/login">Login</a> to vote on this post</p>`
	}

	// Generate comment form
	commentForm := ""
	if currentUser != nil {
		commentForm = fmt.Sprintf(`
		<div class="comment-form">
			<h4>💬 Add a Comment</h4>
			<form method="POST" action="/comments/create">
				<input type="hidden" name="post_id" value="%d">
				<div class="form-group">
					<textarea name="content" required rows="4" placeholder="Share your thoughts..." style="width: 100%%; padding: 12px; border: 2px solid #ddd; border-radius: 6px; font-family: inherit; box-sizing: border-box;"></textarea>
				</div>
				<button type="submit" class="btn btn-success">💬 Post Comment</button>
			</form>
		</div>`, post.ID)
	} else {
		commentForm = `<p><a href="/login">Login</a> to add comments</p>`
	}

	// Generate comments HTML
	commentsHTML := ""
	for _, comment := range comments {
		commentVoteButtons := ""
		if currentUser != nil {
			likeActive := ""
			dislikeActive := ""
			if comment.UserVote != nil {
				if *comment.UserVote {
					likeActive = "active"
				} else {
					dislikeActive = "active"
				}
			}

			commentVoteButtons = fmt.Sprintf(`
			<div class="comment-votes">
				<form method="POST" action="/vote" style="display: inline;">
					<input type="hidden" name="type" value="like">
					<input type="hidden" name="target" value="comment">
					<input type="hidden" name="target_id" value="%d">
					<input type="hidden" name="redirect" value="/posts/view?id=%d">
					<button type="submit" class="btn-small vote-btn like %s">👍 %d</button>
				</form>
				<form method="POST" action="/vote" style="display: inline;">
					<input type="hidden" name="type" value="dislike">
					<input type="hidden" name="target" value="comment">
					<input type="hidden" name="target_id" value="%d">
					<input type="hidden" name="redirect" value="/posts/view?id=%d">
					<button type="submit" class="btn-small vote-btn dislike %s">👎 %d</button>
				</form>
			</div>`, comment.ID, post.ID, likeActive, comment.LikeCount, comment.ID, post.ID, dislikeActive, comment.DislikeCount)
		}

		commentsHTML += fmt.Sprintf(`
		<div class="comment-card">
			<div class="comment-header">
				<div>
					<strong>👤 %s</strong>
					<span class="comment-meta">• %s</span>
				</div>
				%s
			</div>
			<div class="comment-content">%s</div>
		</div>`, comment.Author.Username, comment.CreatedAt.Format("Jan 2, 2006 15:04"), commentVoteButtons, template.HTMLEscapeString(comment.Content))
	}

	if len(comments) == 0 {
		commentsHTML = `
		<div style="text-align: center; padding: 40px; color: #7f8c8d;">
			<h4>💭 No comments yet</h4>
			<p>Be the first to share your thoughts on this post!</p>
		</div>`
	}

	// Generate categories display
	categoriesHTML := ""
	for i, cat := range post.Categories {
		if i > 0 {
			categoriesHTML += ", "
		}
		categoriesHTML += cat.Name
	}
	if categoriesHTML == "" {
		categoriesHTML = "No categories"
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>%s - Forum</title>
		<style>
			%s
			.post-content { font-size: 1.1em; line-height: 1.8; white-space: pre-wrap; }
			.post-meta { color: #7f8c8d; margin: 15px 0; }
			.vote-section { display: flex; align-items: center; gap: 20px; margin: 20px 0; 
							padding: 15px; background: #f8f9fa; border-radius: 6px; }
			.vote-buttons { display: flex; gap: 10px; }
			.vote-btn { padding: 8px 16px; font-size: 0.9em; }
			.vote-btn.like { background: #e8f5e8; color: #28a745; }
			.vote-btn.dislike { background: #f8d7da; color: #dc3545; }
			.vote-btn.active.like { background: #28a745; color: white; }
			.vote-btn.active.dislike { background: #dc3545; color: white; }
			.comment-form { background: #e8f4fd; padding: 20px; border-radius: 8px; margin: 20px 0; }
			.comment-card { background: #f8f9fa; padding: 20px; margin: 15px 0; 
							border-radius: 8px; border-left: 4px solid #007bff; }
			.comment-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
			.comment-meta { color: #7f8c8d; font-size: 0.9em; }
			.comment-content { margin: 10px 0; }
			.comment-votes { display: flex; gap: 10px; }
			.btn-small { padding: 4px 8px; font-size: 0.8em; border: none; border-radius: 4px; cursor: pointer; }
			.form-group { margin-bottom: 15px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>%s</h1>
			</div>
			
			<div class="card">
				<div class="post-meta">
					<p><strong>👤 Author:</strong> %s</p>
					<p><strong>📅 Posted:</strong> %s</p>
					<p><strong>🏷️ Categories:</strong> %s</p>
				</div>
				
				<div class="post-content">%s</div>
				
				<div class="vote-section">
					<div class="vote-stats">
						<strong>👍 %d Likes • 👎 %d Dislikes</strong>
					</div>
					%s
				</div>
			</div>
			
			<div class="card">
				<h3>💬 Comments (%d)</h3>
				%s
				%s
			</div>
			
			<div>
				<a href="/posts" class="btn btn-secondary">← Back to All Posts</a>
				<a href="/" class="btn btn-primary">🏠 Home</a>
			</div>
		</div>
	</body>
	</html>`,
		template.HTMLEscapeString(post.Title), getCommonCSS(),
		template.HTMLEscapeString(post.Title),
		template.HTMLEscapeString(post.Author.Username),
		post.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		template.HTMLEscapeString(categoriesHTML),
		template.HTMLEscapeString(post.Content),
		post.LikeCount, post.DislikeCount, voteButtons,
		len(comments), commentForm, commentsHTML)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// Helper functions
func (h *PostsHandler) showError(w http.ResponseWriter, message string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>Error</title><style>%s</style></head>
	<body>
		<div class="container">
			<div class="error-container">
				<h1>❌ Error</h1>
				<p>%s</p>
				<a href="javascript:history.back()" class="btn btn-secondary">← Go Back</a>
				<a href="/" class="btn btn-primary">🏠 Home</a>
			</div>
		</div>
	</body>
	</html>`, getCommonCSS(), template.HTMLEscapeString(message))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, html)
}

func (h *PostsHandler) showCreatePostSuccess(w http.ResponseWriter, postID int64, title string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head><title>Post Created</title><style>%s</style></head>
	<body>
		<div class="container">
			<div class="card" style="text-align: center;">
				<h1>🎉 Post Created Successfully!</h1>
				<p>Your post "<strong>%s</strong>" has been published!</p>
				<p><strong>Post ID:</strong> #%d</p>
				<div style="margin-top: 30px;">
					<a href="/posts/view?id=%d" class="btn btn-primary">👁️ View Your Post</a>
					<a href="/posts" class="btn btn-secondary">📝 All Posts</a>
					<a href="/posts/create" class="btn btn-success">➕ Create Another Post</a>
				</div>
			</div>
		</div>
	</body>
	</html>`, getCommonCSS(), template.HTMLEscapeString(title), postID, postID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func getActiveClass(isActive bool) string {
	if isActive {
		return "btn-primary active"
	}
	return "btn-secondary"
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
