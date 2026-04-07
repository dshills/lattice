CREATE TABLE IF NOT EXISTS work_item_tags (
    item_id CHAR(36)     NOT NULL,
    tag     VARCHAR(100) NOT NULL,
    PRIMARY KEY (item_id, tag),
    CONSTRAINT fk_tags_item FOREIGN KEY (item_id) REFERENCES work_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
