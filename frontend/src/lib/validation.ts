import { z } from "zod/v4";

const relationshipTypes = [
  "depends_on",
  "blocks",
  "relates_to",
  "duplicate_of",
] as const;

export const createWorkItemSchema = z.object({
  title: z.string().min(1, "Title is required").max(500, "Title max 500 chars"),
  description: z.string().max(10000, "Description max 10000 chars").optional(),
  type: z.string().optional(),
  tags: z.array(z.string()).optional(),
  parent_id: z.string().uuid("Invalid parent ID").optional(),
});

export const updateWorkItemSchema = z.object({
  title: z
    .string()
    .min(1, "Title is required")
    .max(500, "Title max 500 chars")
    .optional(),
  description: z
    .string()
    .max(10000, "Description max 10000 chars")
    .optional(),
  state: z.enum(["NotDone", "InProgress", "Completed"]).optional(),
  type: z.string().optional(),
  tags: z.array(z.string()).optional(),
  parent_id: z.string().optional(),
  override: z.boolean().optional(),
});

export const addRelationshipSchema = z.object({
  type: z.enum(relationshipTypes),
  target_id: z.string().uuid("Invalid target ID"),
});

export type CreateWorkItemData = z.infer<typeof createWorkItemSchema>;
export type UpdateWorkItemData = z.infer<typeof updateWorkItemSchema>;
export type AddRelationshipData = z.infer<typeof addRelationshipSchema>;
