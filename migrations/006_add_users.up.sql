-- Create users table
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(72) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uq_users_email (email)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create project memberships table
CREATE TABLE project_memberships (
    id CHAR(36) PRIMARY KEY,
    project_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    role ENUM('owner', 'member', 'viewer') NOT NULL DEFAULT 'member',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_membership (project_id, user_id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Add assignee_id and created_by columns to work_items
ALTER TABLE work_items ADD COLUMN assignee_id CHAR(36) NULL;
ALTER TABLE work_items ADD COLUMN created_by CHAR(36) NULL;
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE work_items ADD INDEX idx_work_items_assignee (assignee_id);
