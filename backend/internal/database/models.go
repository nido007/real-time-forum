package database

import (
	"time"
)

// User represents a forum user account
// This struct maps to the 'users' table in the database
type User struct {
	ID           int       `json:"id" db:"id"`                 // Primary key - unique user identifier
	Username     string    `json:"username" db:"username"`     // Unique username for login and display
	Email        string    `json:"email" db:"email"`           // Unique email address for login
	PasswordHash string    `json:"-" db:"password_hash"`       // Hashed password (never send in JSON)
	CreatedAt    time.Time `json:"created_at" db:"created_at"` // When the user account was created
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"` // When the user account was last updated
}

// Session represents a user login session
// This struct maps to the 'sessions' table in the database
type Session struct {
	ID        int       `json:"id" db:"id"`                 // Primary key - unique session identifier
	UserID    int       `json:"user_id" db:"user_id"`       // Foreign key to users table
	Token     string    `json:"token" db:"token"`           // Unique session token (stored in cookie)
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"` // When this session expires
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When this session was created

	// Related data - not stored in database but populated when needed
	User *User `json:"user,omitempty" db:"-"` // User associated with this session
}

// Category represents a post category/topic
// This struct maps to the 'categories' table in the database
type Category struct {
	ID          int       `json:"id" db:"id"`                   // Primary key - unique category identifier
	Name        string    `json:"name" db:"name"`               // Category name (e.g., "Technology", "Gaming")
	Description string    `json:"description" db:"description"` // Brief description of the category
	CreatedAt   time.Time `json:"created_at" db:"created_at"`   // When this category was created

	// Related data - not stored in database but populated when needed
	PostCount int `json:"post_count,omitempty" db:"-"` // Number of posts in this category
}

// Post represents a forum post
// This struct maps to the 'posts' table in the database
type Post struct {
	ID        int       `json:"id" db:"id"`                 // Primary key - unique post identifier
	UserID    int       `json:"user_id" db:"user_id"`       // Foreign key to users table (post author)
	Title     string    `json:"title" db:"title"`           // Post title/subject
	Content   string    `json:"content" db:"content"`       // Post content/body
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When this post was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // When this post was last updated

	// Related data - not stored in database but populated when needed
	Author       *User      `json:"author,omitempty" db:"-"`        // User who created this post
	Categories   []Category `json:"categories,omitempty" db:"-"`    // Categories this post belongs to
	Comments     []Comment  `json:"comments,omitempty" db:"-"`      // Comments on this post
	LikeCount    int        `json:"like_count,omitempty" db:"-"`    // Number of likes this post has
	DislikeCount int        `json:"dislike_count,omitempty" db:"-"` // Number of dislikes this post has
	UserVote     *bool      `json:"user_vote,omitempty" db:"-"`     // Current user's vote (true=like, false=dislike, nil=no vote)
	NetScore     int        `json:"net_score,omitempty" db:"-"`     // Likes minus dislikes
	CommentCount int        `json:"comment_count,omitempty" db:"-"` // Total number of comments
}

// PostCategory represents the many-to-many relationship between posts and categories
// This struct maps to the 'post_categories' table in the database
type PostCategory struct {
	PostID     int       `json:"post_id" db:"post_id"`         // Foreign key to posts table
	CategoryID int       `json:"category_id" db:"category_id"` // Foreign key to categories table
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // When this relationship was created

	// Related data - not stored in database but populated when needed
	Post     *Post     `json:"post,omitempty" db:"-"`     // Post associated with this relationship
	Category *Category `json:"category,omitempty" db:"-"` // Category associated with this relationship
}

// Comment represents a comment on a forum post
// This struct maps to the 'comments' table in the database
type Comment struct {
	ID        int       `json:"id" db:"id"`                 // Primary key - unique comment identifier
	PostID    int       `json:"post_id" db:"post_id"`       // Foreign key to posts table
	UserID    int       `json:"user_id" db:"user_id"`       // Foreign key to users table (comment author)
	Content   string    `json:"content" db:"content"`       // Comment content/text
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When this comment was created
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // When this comment was last updated

	// Related data - not stored in database but populated when needed
	Author       *User `json:"author,omitempty" db:"-"`        // User who created this comment
	Post         *Post `json:"post,omitempty" db:"-"`          // Post this comment belongs to
	LikeCount    int   `json:"like_count,omitempty" db:"-"`    // Number of likes this comment has
	DislikeCount int   `json:"dislike_count,omitempty" db:"-"` // Number of dislikes this comment has
	UserVote     *bool `json:"user_vote,omitempty" db:"-"`     // Current user's vote on this comment
	NetScore     int   `json:"net_score,omitempty" db:"-"`     // Likes minus dislikes
}

// Like represents a like or dislike vote on a post or comment
// This struct maps to the 'likes' table in the database
type Like struct {
	ID        int       `json:"id" db:"id"`                 // Primary key - unique like identifier
	UserID    int       `json:"user_id" db:"user_id"`       // Foreign key to users table (who voted)
	PostID    *int      `json:"post_id" db:"post_id"`       // Foreign key to posts table (if voting on post)
	CommentID *int      `json:"comment_id" db:"comment_id"` // Foreign key to comments table (if voting on comment)
	IsLike    bool      `json:"is_like" db:"is_like"`       // true = like, false = dislike
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When this vote was cast

	// Related data - not stored in database but populated when needed
	User    *User    `json:"user,omitempty" db:"-"`    // User who cast this vote
	Post    *Post    `json:"post,omitempty" db:"-"`    // Post being voted on (if applicable)
	Comment *Comment `json:"comment,omitempty" db:"-"` // Comment being voted on (if applicable)
}

// ContactMessage represents a message sent through the contact form
// This struct maps to the 'contact_messages' table in the database
type ContactMessage struct {
	ID        int       `json:"id" db:"id"`                 // Primary key - unique message identifier
	Name      string    `json:"name" db:"name"`             // Name of the person sending the message
	Email     string    `json:"email" db:"email"`           // Email address of the sender
	Message   string    `json:"message" db:"message"`       // The actual message content
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When this message was sent
}

// PostFilter represents filters for querying posts
// This struct is used for filtering posts by various criteria
type PostFilter struct {
	UserID     *int   `json:"user_id,omitempty"`     // Filter by specific user's posts
	CategoryID *int   `json:"category_id,omitempty"` // Filter by specific category
	LikedBy    *int   `json:"liked_by,omitempty"`    // Filter by posts liked by specific user
	DislikedBy *int   `json:"disliked_by,omitempty"` // Filter by posts disliked by specific user
	Search     string `json:"search,omitempty"`      // Search in title and content
	Limit      int    `json:"limit,omitempty"`       // Maximum number of posts to return
	Offset     int    `json:"offset,omitempty"`      // Number of posts to skip (for pagination)
	SortBy     string `json:"sort_by,omitempty"`     // Sort field (created_at, title, likes, etc.)
	SortOrder  string `json:"sort_order,omitempty"`  // Sort direction (ASC, DESC)
	MinLikes   *int   `json:"min_likes,omitempty"`   // Minimum number of likes
	MaxAge     *int   `json:"max_age,omitempty"`     // Maximum age in days
}

// VoteStats represents aggregated voting statistics
// This struct is used for displaying vote counts and user voting status
type VoteStats struct {
	LikeCount    int   `json:"like_count"`          // Total number of likes
	DislikeCount int   `json:"dislike_count"`       // Total number of dislikes
	UserVote     *bool `json:"user_vote,omitempty"` // Current user's vote (if any)
	NetScore     int   `json:"net_score"`           // Likes minus dislikes
	TotalVotes   int   `json:"total_votes"`         // Total votes (likes + dislikes)
}

// UserStats represents user activity statistics
// This struct is used for displaying user profile information
type UserStats struct {
	User             *User      `json:"user"`                  // User information
	PostCount        int        `json:"post_count"`            // Number of posts created by user
	CommentCount     int        `json:"comment_count"`         // Number of comments created by user
	LikesGiven       int        `json:"likes_given"`           // Number of likes given by user
	DislikesGiven    int        `json:"dislikes_given"`        // Number of dislikes given by user
	LikesReceived    int        `json:"likes_received"`        // Number of likes received by user
	DislikesReceived int        `json:"dislikes_received"`     // Number of dislikes received by user
	NetKarma         int        `json:"net_karma"`             // Net karma (likes received - dislikes received)
	JoinedDays       int        `json:"joined_days"`           // Days since user joined
	LastActive       *time.Time `json:"last_active,omitempty"` // When user was last active
}

// CategoryStats represents category activity statistics
// This struct is used for displaying category information with activity data
type CategoryStats struct {
	Category     *Category  `json:"category"`               // Category information
	PostCount    int        `json:"post_count"`             // Number of posts in this category
	CommentCount int        `json:"comment_count"`          // Number of comments in posts of this category
	LastPostAt   *time.Time `json:"last_post_at,omitempty"` // When the most recent post was made
	ActiveUsers  int        `json:"active_users"`           // Number of users who posted in this category
	TotalLikes   int        `json:"total_likes"`            // Total likes in this category
	TotalVotes   int        `json:"total_votes"`            // Total votes in this category
}

// ForumStats represents overall forum statistics
// This struct is used for displaying general forum activity and health
type ForumStats struct {
	TotalUsers      int       `json:"total_users"`      // Total registered users
	TotalPosts      int       `json:"total_posts"`      // Total posts created
	TotalComments   int       `json:"total_comments"`   // Total comments created
	TotalCategories int       `json:"total_categories"` // Total categories available
	TotalVotes      int       `json:"total_votes"`      // Total votes (likes + dislikes)
	TotalLikes      int       `json:"total_likes"`      // Total likes given
	TotalDislikes   int       `json:"total_dislikes"`   // Total dislikes given
	ActiveUsers24h  int       `json:"active_users_24h"` // Users active in last 24 hours
	PostsToday      int       `json:"posts_today"`      // Posts created today
	CommentsToday   int       `json:"comments_today"`   // Comments created today
	LastActivity    time.Time `json:"last_activity"`    // Most recent activity timestamp
}

// ValidationError represents a data validation error
// This struct is used for handling and displaying validation errors
type ValidationError struct {
	Field   string `json:"field"`   // Name of the field that failed validation
	Message string `json:"message"` // Human-readable error message
	Code    string `json:"code"`    // Error code for programmatic handling
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	return ve.Message
}

// PaginationInfo represents pagination information for lists
// This struct is used for paginating long lists of posts, comments, etc.
type PaginationInfo struct {
	CurrentPage  int  `json:"current_page"`   // Current page number (1-based)
	TotalPages   int  `json:"total_pages"`    // Total number of pages
	TotalItems   int  `json:"total_items"`    // Total number of items
	ItemsPerPage int  `json:"items_per_page"` // Number of items per page
	HasNext      bool `json:"has_next"`       // Whether there's a next page
	HasPrev      bool `json:"has_prev"`       // Whether there's a previous page
	StartItem    int  `json:"start_item"`     // Index of first item on current page
	EndItem      int  `json:"end_item"`       // Index of last item on current page
}

// SearchResult represents a search result item
// This struct is used for search functionality across posts and comments
type SearchResult struct {
	Type      string    `json:"type"`              // "post" or "comment"
	ID        int       `json:"id"`                // ID of the post or comment
	Title     string    `json:"title"`             // Title (for posts) or excerpt (for comments)
	Content   string    `json:"content"`           // Content excerpt with highlighted terms
	Author    string    `json:"author"`            // Author username
	CreatedAt time.Time `json:"created_at"`        // When it was created
	Relevance float64   `json:"relevance"`         // Search relevance score
	URL       string    `json:"url"`               // Direct URL to the content
	PostID    *int      `json:"post_id,omitempty"` // Parent post ID (for comments)
}

// ActivityLog represents user activity logging
// This struct is used for tracking user actions for security and analytics
type ActivityLog struct {
	ID         int       `json:"id" db:"id"`                   // Primary key
	UserID     *int      `json:"user_id" db:"user_id"`         // User ID (null for anonymous actions)
	Action     string    `json:"action" db:"action"`           // Action performed (login, post_create, etc.)
	EntityType string    `json:"entity_type" db:"entity_type"` // Type of entity affected (post, comment, user)
	EntityID   *int      `json:"entity_id" db:"entity_id"`     // ID of the affected entity
	IPAddress  string    `json:"ip_address" db:"ip_address"`   // IP address of the user
	UserAgent  string    `json:"user_agent" db:"user_agent"`   // Browser user agent
	Details    string    `json:"details" db:"details"`         // Additional details in JSON format
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // When the action occurred

	// Related data
	User *User `json:"user,omitempty" db:"-"` // User who performed the action
}

// NotificationPreferences represents user notification settings
// This struct is used for managing user notification preferences
type NotificationPreferences struct {
	UserID            int  `json:"user_id" db:"user_id"`                       // Foreign key to users table
	EmailOnComment    bool `json:"email_on_comment" db:"email_on_comment"`     // Email when someone comments on user's post
	EmailOnLike       bool `json:"email_on_like" db:"email_on_like"`           // Email when someone likes user's content
	EmailOnReply      bool `json:"email_on_reply" db:"email_on_reply"`         // Email when someone replies to user's comment
	EmailDigest       bool `json:"email_digest" db:"email_digest"`             // Weekly digest email
	PushNotifications bool `json:"push_notifications" db:"push_notifications"` // Browser push notifications

	// Related data
	User *User `json:"user,omitempty" db:"-"` // User these preferences belong to
}

// FileUpload represents uploaded files (for future file attachment feature)
// This struct can be used for implementing file attachments to posts
type FileUpload struct {
	ID        int       `json:"id" db:"id"`                 // Primary key
	UserID    int       `json:"user_id" db:"user_id"`       // User who uploaded the file
	PostID    *int      `json:"post_id" db:"post_id"`       // Post this file is attached to (optional)
	CommentID *int      `json:"comment_id" db:"comment_id"` // Comment this file is attached to (optional)
	Filename  string    `json:"filename" db:"filename"`     // Original filename
	StorePath string    `json:"store_path" db:"store_path"` // Path where file is stored
	FileSize  int64     `json:"file_size" db:"file_size"`   // File size in bytes
	MimeType  string    `json:"mime_type" db:"mime_type"`   // MIME type of the file
	CreatedAt time.Time `json:"created_at" db:"created_at"` // When file was uploaded

	// Related data
	User    *User    `json:"user,omitempty" db:"-"`    // User who uploaded
	Post    *Post    `json:"post,omitempty" db:"-"`    // Associated post
	Comment *Comment `json:"comment,omitempty" db:"-"` // Associated comment
}

// Tag represents content tags (for future tagging feature)
// This struct can be used for implementing a tagging system
type Tag struct {
	ID          int       `json:"id" db:"id"`                   // Primary key
	Name        string    `json:"name" db:"name"`               // Tag name (e.g., "golang", "beginner")
	Description string    `json:"description" db:"description"` // Tag description
	Color       string    `json:"color" db:"color"`             // Display color (hex code)
	CreatedAt   time.Time `json:"created_at" db:"created_at"`   // When tag was created

	// Related data
	PostCount int `json:"post_count,omitempty" db:"-"` // Number of posts with this tag
}
