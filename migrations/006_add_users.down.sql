-- Remove assignee and creator columns from work_items
ALTER TABLE work_items DROP INDEX idx_work_items_assignee;
ALTER TABLE work_items DROP FOREIGN KEY fk_work_items_creator;
ALTER TABLE work_items DROP FOREIGN KEY fk_work_items_assignee;
ALTER TABLE work_items DROP COLUMN created_by;
ALTER TABLE work_items DROP COLUMN assignee_id;

-- Drop project memberships table
DROP TABLE IF EXISTS project_memberships;

-- Drop users table
DROP TABLE IF EXISTS users;
