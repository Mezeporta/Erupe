package channelserver

import (
	"context"
	"time"
)

// ListPosts returns active guild posts of the given type, ordered by newest first.
func (r *GuildRepository) ListPosts(guildID uint32, postType int) ([]*MessageBoardPost, error) {
	rows, err := r.db.Queryx(
		`SELECT id, stamp_id, title, body, author_id, created_at, liked_by
		 FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false
		 ORDER BY created_at DESC`, guildID, postType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var posts []*MessageBoardPost
	for rows.Next() {
		post := &MessageBoardPost{}
		if err := rows.StructScan(post); err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// CreatePost inserts a new guild post and soft-deletes excess posts beyond maxPosts.
func (r *GuildRepository) CreatePost(guildID, authorID, stampID uint32, postType int, title, body string, maxPosts int) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`INSERT INTO guild_posts (guild_id, author_id, stamp_id, post_type, title, body) VALUES ($1, $2, $3, $4, $5, $6)`,
		guildID, authorID, stampID, postType, title, body); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE guild_posts SET deleted = true WHERE id IN (
		SELECT id FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false
		ORDER BY created_at DESC OFFSET $3
	)`, guildID, postType, maxPosts); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePost soft-deletes a guild post by ID.
func (r *GuildRepository) DeletePost(postID uint32) error {
	_, err := r.db.Exec("UPDATE guild_posts SET deleted = true WHERE id = $1", postID)
	return err
}

// UpdatePost updates the title and body of a guild post.
func (r *GuildRepository) UpdatePost(postID uint32, title, body string) error {
	_, err := r.db.Exec("UPDATE guild_posts SET title = $1, body = $2 WHERE id = $3", title, body, postID)
	return err
}

// UpdatePostStamp updates the stamp of a guild post.
func (r *GuildRepository) UpdatePostStamp(postID, stampID uint32) error {
	_, err := r.db.Exec("UPDATE guild_posts SET stamp_id = $1 WHERE id = $2", stampID, postID)
	return err
}

// GetPostLikedBy returns the liked_by CSV string for a guild post.
func (r *GuildRepository) GetPostLikedBy(postID uint32) (string, error) {
	var likedBy string
	err := r.db.QueryRow("SELECT liked_by FROM guild_posts WHERE id = $1", postID).Scan(&likedBy)
	return likedBy, err
}

// SetPostLikedBy updates the liked_by CSV string for a guild post.
func (r *GuildRepository) SetPostLikedBy(postID uint32, likedBy string) error {
	_, err := r.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, postID)
	return err
}

// CountNewPosts returns the count of non-deleted posts created after the given time.
func (r *GuildRepository) CountNewPosts(guildID uint32, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM guild_posts WHERE guild_id = $1 AND deleted = false AND (EXTRACT(epoch FROM created_at)::int) > $2`,
		guildID, since.Unix()).Scan(&count)
	return count, err
}
