import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listProjects,
  getProject,
  createProject,
  updateProject,
  deleteProject,
} from "../lib/api/projects";
import { useToast } from "../components/common/Toast";
import { toastError } from "../lib/toastError";
import type { CreateProjectInput, UpdateProjectInput } from "../lib/types";

export function useProjects() {
  return useQuery({
    queryKey: ["projects"],
    queryFn: () => listProjects(),
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: ["project", id],
    queryFn: () => getProject(id),
    enabled: !!id,
  });
}

export function useProjectMutations() {
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["projects"] });
  };

  const createMutation = useMutation({
    mutationFn: (input: CreateProjectInput) => createProject(input),
    onSuccess: () => {
      invalidate();
      addToast("Project created", "success");
    },
    onError: (err) => toastError(addToast, err),
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      input,
    }: {
      id: string;
      input: UpdateProjectInput;
    }) => updateProject(id, input),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["project", variables.id],
      });
      invalidate();
    },
    onError: (err) => toastError(addToast, err),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => {
      invalidate();
      addToast("Project deleted", "success");
    },
    onError: (err) => toastError(addToast, err),
  });

  return { createMutation, updateMutation, deleteMutation };
}
