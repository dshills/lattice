import { ApiClientError } from "./api/client";

export function toastError(addToast: (msg: string, type?: "error") => void, err: unknown) {
  if (err instanceof ApiClientError) {
    if (err.status === 403) {
      addToast("Permission denied", "error");
    } else if (err.status === 409) {
      addToast("Conflict: " + err.message, "error");
    } else if (err.status >= 500) {
      addToast("Server error — please try again", "error");
    } else {
      addToast(err.message, "error");
    }
  } else {
    addToast(err instanceof Error ? err.message : "An unexpected error occurred", "error");
  }
}
