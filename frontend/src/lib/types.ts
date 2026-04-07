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
  title: string;
  description: string;
  state: WorkItemState;
  tags: string[];
  type: string;
  parent_id: string | null;
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
  override?: boolean;
}

export interface AddRelationshipInput {
  type: RelationshipType;
  target_id: string;
}
