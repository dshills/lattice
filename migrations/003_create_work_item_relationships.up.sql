CREATE TABLE IF NOT EXISTS work_item_relationships (
    id        CHAR(36)    NOT NULL PRIMARY KEY,
    source_id CHAR(36)    NOT NULL,
    target_id CHAR(36)    NOT NULL,
    type      VARCHAR(50) NOT NULL,
    UNIQUE KEY uk_source_target_type (source_id, target_id, type),
    CONSTRAINT fk_rel_source FOREIGN KEY (source_id) REFERENCES work_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_rel_target FOREIGN KEY (target_id) REFERENCES work_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
