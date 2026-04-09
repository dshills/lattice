-- Create projects table
CREATE TABLE projects (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uq_projects_name (name)
);

-- Create default project for existing data
INSERT INTO projects (id, name, description)
VALUES ('00000000-0000-0000-0000-000000000001', 'Default', 'Auto-created project for existing work items');

-- Add project_id to work_items, backfill, enforce NOT NULL + FK
ALTER TABLE work_items ADD COLUMN project_id CHAR(36) NULL AFTER id;
UPDATE work_items SET project_id = '00000000-0000-0000-0000-000000000001' WHERE project_id IS NULL;
ALTER TABLE work_items MODIFY COLUMN project_id CHAR(36) NOT NULL;
ALTER TABLE work_items ADD INDEX idx_work_items_project_id (project_id);
ALTER TABLE work_items ADD CONSTRAINT fk_work_items_project FOREIGN KEY (project_id) REFERENCES projects(id);
