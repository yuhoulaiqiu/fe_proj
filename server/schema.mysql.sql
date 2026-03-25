CREATE DATABASE IF NOT EXISTS community_help_hub DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE community_help_hub;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT NOT NULL AUTO_INCREMENT,
  username VARCHAR(191) NOT NULL,
  password VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  role VARCHAR(32) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sessions (
  id BIGINT NOT NULL AUTO_INCREMENT,
  token VARCHAR(255) NOT NULL,
  user_id BIGINT NOT NULL,
  expires_at VARCHAR(64) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_sessions_token (token),
  KEY idx_sessions_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS activities (
  id BIGINT NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  category VARCHAR(64) NOT NULL DEFAULT '其他',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  user_id BIGINT NOT NULL DEFAULT 0,
  cover_url TEXT NOT NULL,
  summary TEXT NOT NULL,
  content MEDIUMTEXT NOT NULL,
  location VARCHAR(255) NOT NULL,
  start_time VARCHAR(64) NOT NULL,
  end_time VARCHAR(64) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  deleted_at VARCHAR(64) NULL,
  PRIMARY KEY (id),
  KEY idx_activities_category_status (category, status),
  KEY idx_activities_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS activity_registrations (
  id BIGINT NOT NULL AUTO_INCREMENT,
  activity_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_activity_user (activity_id, user_id),
  KEY idx_registrations_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS services (
  id BIGINT NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  category VARCHAR(64) NOT NULL,
  phone VARCHAR(64) NOT NULL,
  address VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_services_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS lost_items (
  id BIGINT NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  item_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  location VARCHAR(255) NOT NULL,
  occurred_at VARCHAR(64) NOT NULL,
  description TEXT NOT NULL,
  contact VARCHAR(255) NOT NULL,
  created_at VARCHAR(64) NOT NULL,
  updated_at VARCHAR(64) NOT NULL,
  deleted_at VARCHAR(64) NULL,
  PRIMARY KEY (id),
  KEY idx_lost_items_deleted_at (deleted_at),
  KEY idx_lost_items_type_status (item_type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT NOT NULL AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  kind VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  activity_id BIGINT NULL,
  scheduled_for VARCHAR(64) NOT NULL,
  read_at VARCHAR(64) NULL,
  created_at VARCHAR(64) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_notifications_user_activity_kind (user_id, activity_id, kind),
  KEY idx_notifications_user_scheduled (user_id, scheduled_for),
  KEY idx_notifications_read_at (read_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
