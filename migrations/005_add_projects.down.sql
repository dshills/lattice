ALTER TABLE work_items DROP FOREIGN KEY fk_work_items_project;
ALTER TABLE work_items DROP INDEX idx_work_items_project_id;
ALTER TABLE work_items DROP COLUMN project_id;
DROP TABLE projects;
