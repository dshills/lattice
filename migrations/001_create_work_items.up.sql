CREATE TABLE IF NOT EXISTS work_items (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    title      VARCHAR(500) NOT NULL,
    description TEXT        NOT NULL,
    state      VARCHAR(20)  NOT NULL DEFAULT 'NotDone',
    type       VARCHAR(100) DEFAULT NULL,
    parent_id  CHAR(36)     DEFAULT NULL,
    created_at DATETIME(3)  NOT NULL,
    updated_at DATETIME(3)  NOT NULL,
    INDEX idx_state (state),
    INDEX idx_type (type),
    INDEX idx_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
