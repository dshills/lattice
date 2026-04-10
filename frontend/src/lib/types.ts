export interface User {
  id: string;
  email: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectWithCount extends Project {
  item_count: number;
  role?: string;
}

export interface ProjectListResponse {
  projects: ProjectWithCount[];
}

export interface CreateProjectInput {
  name: string;
  description?: string;
}

export interface UpdateProjectInput {
  name?: string;
  description?: string;
}

export type WorkItemState = "NotDone" | "InProgress" | "Completed";

export type RelationshipType =
  | "depends_on"
  | "blocks"
  | "relates_to"
  | "duplicate_of";

export interface Relationship {
  id: string;
  type: RelationshipType;
  target_id: string;
}

export interface WorkItem {
  id: string;
  project_id: string;
  title: string;
  description: string;
  state: WorkItemState;
  tags: string[];
  type: string;
  parent_id: string | null;
  assignee_id: string | null;
  created_by: string | null;
  assignee_name?: string;
  relationships: Relationship[];
  created_at: string;
  updated_at: string;
  is_blocked: boolean;
}

export interface ListResponse {
  items: WorkItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface ApiError {
  error: {
    code: string;
    message: string;
  };
}

export interface ListFilter {
  state?: WorkItemState;
  tags?: string;
  type?: string;
  parent_id?: string;
  assignee_id?: string;
  is_blocked?: boolean;
  is_ready?: boolean;
  page?: number;
  page_size?: number;
}

export interface CreateWorkItemInput {
  title: string;
  description?: string;
  type?: string;
  tags?: string[];
  parent_id?: string;
}

export interface UpdateWorkItemInput {
  title?: string;
  description?: string;
  state?: WorkItemState;
  type?: string;
  tags?: string[];
  parent_id?: string;
  assignee_id?: string;
  override?: boolean;
}

export interface AddRelationshipInput {
  type: RelationshipType;
  target_id: string;
}
