import { writable, derived } from "svelte/store";
import type { V1User } from "../../client";

export interface ViewAsUserState {
  user: V1User;
  /**
   * The project where "View As" was activated.
   * Used for querying users in the dropdown.
   */
  sourceProject: string;
  /**
   * If true, the user is an org-level admin and the view-as persists across all projects.
   * If false, the view-as is scoped to the sourceProject only.
   */
  isOrgLevel: boolean;
}

const viewAsUserStateStore = writable<ViewAsUserState | null>(null);

/**
 * Sets the "View As" user with the project context.
 * @param user The user to view as
 * @param sourceProject The project where this was activated (used for querying users)
 * @param isOrgLevel Whether this is an org-level view-as that persists across projects
 */
export function setViewAsUser(
  user: V1User,
  sourceProject: string,
  isOrgLevel: boolean,
): void {
  viewAsUserStateStore.set({ user, sourceProject, isOrgLevel });
}

/**
 * Clears the "View As" state.
 */
export function clearViewAsUser(): void {
  viewAsUserStateStore.set(null);
}

/**
 * Checks if "View As" is valid for the given project context.
 * Returns true if:
 * - No "View As" is active (nothing to validate)
 * - The "View As" is org-level (persists across all contexts)
 * - The current project matches the source project
 */
export function isViewAsValidForProject(
  state: ViewAsUserState | null,
  currentProject: string | null | undefined,
): boolean {
  if (!state) return true; // No view-as active
  if (state.isOrgLevel) return true; // Org-level persists everywhere
  if (!currentProject) return false; // Project-scoped but no current project
  return state.sourceProject === currentProject;
}

/**
 * The full state store for internal use.
 */
export const viewAsUserStateStore$ = viewAsUserStateStore;

/**
 * Derived store that provides just the user for backward compatibility.
 */
export const viewAsUserStore = derived(
  viewAsUserStateStore,
  ($state) => $state?.user ?? null,
);
