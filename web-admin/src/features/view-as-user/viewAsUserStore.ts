import { writable, derived } from "svelte/store";
import type { V1User } from "../../client";

export interface ViewAsUserState {
  user: V1User;
  /**
   * The project where "View As" was activated.
   * If null, it means the user is an org-level admin and can view as across all projects.
   */
  projectContext: string | null;
}

const viewAsUserStateStore = writable<ViewAsUserState | null>(null);

/**
 * Sets the "View As" user with the project context.
 * @param user The user to view as
 * @param projectContext The project where this was activated. Pass null for org-level admins.
 */
export function setViewAsUser(
  user: V1User,
  projectContext: string | null,
): void {
  viewAsUserStateStore.set({ user, projectContext });
}

/**
 * Clears the "View As" state.
 */
export function clearViewAsUser(): void {
  viewAsUserStateStore.set(null);
}

/**
 * Checks if "View As" is valid for the given project.
 * Returns true if:
 * - No "View As" is active (nothing to validate)
 * - The "View As" was activated at org-level (projectContext is null)
 * - The "View As" was activated for this specific project
 */
export function isViewAsValidForProject(
  state: ViewAsUserState | null,
  currentProject: string | null,
): boolean {
  if (!state) return true; // No view-as active
  if (state.projectContext === null) return true; // Org-level admin
  if (!currentProject) return false; // View-as active but no current project
  return state.projectContext === currentProject;
}

/**
 * The full state store for internal use.
 */
export const viewAsUserStateStore$ = viewAsUserStateStore;

/**
 * Derived store that provides just the user for backward compatibility.
 * @deprecated Use viewAsUserStateStore$ for full context
 */
export const viewAsUserStore = derived(
  viewAsUserStateStore,
  ($state) => $state?.user ?? null,
);
