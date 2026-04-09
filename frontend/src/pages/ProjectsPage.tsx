import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { useProjects, useProjectMutations } from "../hooks/useProjects";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { EmptyState } from "../components/common/EmptyState";
import { Modal } from "../components/common/Modal";
import { ConfirmDialog } from "../components/common/ConfirmDialog";
import type { ProjectWithCount } from "../lib/types";

export function ProjectsPage() {
  const navigate = useNavigate();
  const { data, isLoading, error, refetch } = useProjects();
  const { createMutation, updateMutation, deleteMutation } =
    useProjectMutations();
  const [createOpen, setCreateOpen] = useState(false);
  const [editProject, setEditProject] = useState<ProjectWithCount | null>(null);
  const [deleteProject, setDeleteProject] = useState<ProjectWithCount | null>(
    null,
  );
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");

  const handleCreate = () => {
    createMutation.mutate(
      { name: formName, description: formDescription },
      {
        onSuccess: (project) => {
          setCreateOpen(false);
          setFormName("");
          setFormDescription("");
          navigate(`/projects/${project.id}/board`);
        },
      },
    );
  };

  const handleUpdate = () => {
    if (!editProject) return;
    updateMutation.mutate(
      {
        id: editProject.id,
        input: { name: formName, description: formDescription },
      },
      {
        onSuccess: () => {
          setEditProject(null);
          setFormName("");
          setFormDescription("");
        },
      },
    );
  };

  const handleDelete = () => {
    if (!deleteProject) return;
    deleteMutation.mutate(deleteProject.id, {
      onSuccess: () => setDeleteProject(null),
    });
  };

  const openEdit = (project: ProjectWithCount) => {
    setEditProject(project);
    setFormName(project.name);
    setFormDescription(project.description);
  };

  if (isLoading) return <LoadingState />;
  if (error)
    return <ErrorState message={error.message} onRetry={() => refetch()} />;

  const projects = data?.projects ?? [];

  return (
    <div className="mx-auto max-w-3xl">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Projects</h1>
        <button
          onClick={() => {
            setFormName("");
            setFormDescription("");
            setCreateOpen(true);
          }}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          New Project
        </button>
      </div>

      {projects.length === 0 ? (
        <EmptyState
          title="No projects yet"
          description="Create your first project to get started"
        />
      ) : (
        <div className="space-y-2">
          {projects.map((project) => (
            <Link
              key={project.id}
              to={`/projects/${project.id}/board`}
              className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4 hover:border-gray-300"
            >
              <div>
                <h2 className="font-medium text-gray-900">{project.name}</h2>
                {project.description && (
                  <p className="mt-0.5 text-sm text-gray-500 line-clamp-1">
                    {project.description}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm text-gray-400">
                  {project.item_count} item{project.item_count !== 1 ? "s" : ""}
                </span>
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    openEdit(project);
                  }}
                  className="rounded px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
                >
                  Edit
                </button>
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    setDeleteProject(project);
                  }}
                  className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            </Link>
          ))}
        </div>
      )}

      {/* Create modal */}
      <Modal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="New Project"
      >
        <div className="space-y-4">
          <input
            type="text"
            value={formName}
            onChange={(e) => setFormName(e.target.value)}
            placeholder="Project name"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            autoFocus
          />
          <textarea
            value={formDescription}
            onChange={(e) => setFormDescription(e.target.value)}
            placeholder="Description (optional)"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 min-h-[80px]"
          />
          <div className="flex justify-end gap-3">
            <button
              onClick={() => setCreateOpen(false)}
              className="rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
            >
              Cancel
            </button>
            <button
              onClick={handleCreate}
              disabled={!formName.trim() || createMutation.isPending}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {createMutation.isPending ? "Creating..." : "Create"}
            </button>
          </div>
        </div>
      </Modal>

      {/* Edit modal */}
      <Modal
        open={!!editProject}
        onClose={() => setEditProject(null)}
        title="Edit Project"
      >
        <div className="space-y-4">
          <input
            type="text"
            value={formName}
            onChange={(e) => setFormName(e.target.value)}
            placeholder="Project name"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            autoFocus
          />
          <textarea
            value={formDescription}
            onChange={(e) => setFormDescription(e.target.value)}
            placeholder="Description (optional)"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 min-h-[80px]"
          />
          <div className="flex justify-end gap-3">
            <button
              onClick={() => setEditProject(null)}
              className="rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
            >
              Cancel
            </button>
            <button
              onClick={handleUpdate}
              disabled={!formName.trim() || updateMutation.isPending}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {updateMutation.isPending ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      </Modal>

      {/* Delete confirm */}
      <ConfirmDialog
        open={!!deleteProject}
        onClose={() => setDeleteProject(null)}
        onConfirm={handleDelete}
        title="Delete project"
        message={`Are you sure you want to delete "${deleteProject?.name}"? All work items in this project will be deleted.`}
        confirmLabel="Delete"
        danger
      />
    </div>
  );
}
