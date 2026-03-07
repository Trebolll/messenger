DROP INDEX IF EXISTS idx_feed_events_user;
DROP INDEX IF EXISTS idx_feed_events_post;
DROP INDEX IF EXISTS idx_wall_att_post_created;
DROP INDEX IF EXISTS idx_wall_likes_post;
DROP INDEX IF EXISTS idx_wall_likes_user;
DROP INDEX IF EXISTS idx_messages_chat;
DROP INDEX IF EXISTS idx_feed_scores_user_post;
DROP INDEX IF EXISTS idx_wall_posts_created;

DROP TABLE IF EXISTS feed_scores;
DROP TABLE IF EXISTS feed_events;
DROP TABLE IF EXISTS feed_preferences;