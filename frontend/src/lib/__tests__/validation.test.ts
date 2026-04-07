import { describe, it, expect } from "vitest";
import {
  createWorkItemSchema,
  updateWorkItemSchema,
  addRelationshipSchema,
} from "../validation";

describe("createWorkItemSchema", () => {
  it("accepts valid input", () => {
    const result = createWorkItemSchema.safeParse({ title: "Fix bug" });
    expect(result.success).toBe(true);
  });

  it("requires title", () => {
    const result = createWorkItemSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it("rejects empty title", () => {
    const result = createWorkItemSchema.safeParse({ title: "" });
    expect(result.success).toBe(false);
  });

  it("rejects title over 500 chars", () => {
    const result = createWorkItemSchema.safeParse({ title: "a".repeat(501) });
    expect(result.success).toBe(false);
  });

  it("rejects description over 10000 chars", () => {
    const result = createWorkItemSchema.safeParse({
      title: "ok",
      description: "a".repeat(10001),
    });
    expect(result.success).toBe(false);
  });

  it("accepts description at 10000 chars", () => {
    const result = createWorkItemSchema.safeParse({
      title: "ok",
      description: "a".repeat(10000),
    });
    expect(result.success).toBe(true);
  });
});

describe("updateWorkItemSchema", () => {
  it("accepts empty object (all optional)", () => {
    const result = updateWorkItemSchema.safeParse({});
    expect(result.success).toBe(true);
  });

  it("accepts partial update", () => {
    const result = updateWorkItemSchema.safeParse({
      title: "Updated",
      state: "InProgress",
    });
    expect(result.success).toBe(true);
  });

  it("rejects invalid state", () => {
    const result = updateWorkItemSchema.safeParse({ state: "Done" });
    expect(result.success).toBe(false);
  });

  it("rejects title under 1 char when provided", () => {
    const result = updateWorkItemSchema.safeParse({ title: "" });
    expect(result.success).toBe(false);
  });

  it("rejects description over 10000 chars", () => {
    const result = updateWorkItemSchema.safeParse({
      description: "a".repeat(10001),
    });
    expect(result.success).toBe(false);
  });
});

describe("addRelationshipSchema", () => {
  it("accepts valid relationship", () => {
    const result = addRelationshipSchema.safeParse({
      type: "depends_on",
      target_id: "550e8400-e29b-41d4-a716-446655440000",
    });
    expect(result.success).toBe(true);
  });

  it("rejects invalid type", () => {
    const result = addRelationshipSchema.safeParse({
      type: "invalid",
      target_id: "550e8400-e29b-41d4-a716-446655440000",
    });
    expect(result.success).toBe(false);
  });

  it("rejects invalid UUID", () => {
    const result = addRelationshipSchema.safeParse({
      type: "blocks",
      target_id: "not-a-uuid",
    });
    expect(result.success).toBe(false);
  });

  it("rejects missing target_id", () => {
    const result = addRelationshipSchema.safeParse({ type: "blocks" });
    expect(result.success).toBe(false);
  });
});
