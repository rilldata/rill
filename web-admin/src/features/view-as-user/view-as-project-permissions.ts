import { derived, type Readable } from "svelte/store";
import {
  createAdminServiceGetDeploymentCredentials,
  type V1GetDeploymentCredentialsResponse,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
import { viewAsUserStore } from "./viewAsUserStore";

export interface ViewAsState {
  /** The mocked user's ID, or undefined when View As is inactive */
  mockedUserId: string | undefined;
  /** The mocked user's deployment credentials (for RuntimeProvider) */
  deploymentCredentials: V1GetDeploymentCredentialsResponse | undefined;
  /** The mocked user's project permissions, or undefined when View As is inactive */
  projectPermissions: V1ProjectPermissions | undefined;
  /** Whether the View As queries are currently loading */
  isLoading: boolean;
}

/**
 * Creates a compound store that encapsulates the entire View As query chain:
 * viewAsUserStore → GetDeploymentCredentials → GetProjectWithBearerToken → permissions
 *
 * This hook eliminates duplication between consumers and ensures consistent
 * handling of the mocked user's permissions across the app.
 *
 * TanStack Query deduplicates by query key, so multiple consumers calling this
 * hook get instant cache hits with zero extra network calls.
 *
 * @param org - Organization name
 * @param project - Project name
 * @returns A readable store containing the View As state
 */
export function createViewAsState(
  org: string,
  project: string,
): Readable<ViewAsState> {
  return derived(viewAsUserStore, ($viewAsUser, set) => {
    const mockedUserId = $viewAsUser?.id;

    if (!mockedUserId) {
      set({
        mockedUserId: undefined,
        deploymentCredentials: undefined,
        projectPermissions: undefined,
        isLoading: false,
      });
      return;
    }

    const credentialsQuery = createAdminServiceGetDeploymentCredentials(
      org,
      project,
      { userId: mockedUserId },
      {
        query: {
          enabled: true,
        },
      },
    );

    const unsubCredentials = credentialsQuery.subscribe(($credentialsQuery) => {
      const accessToken = $credentialsQuery.data?.accessToken;

      if (!accessToken) {
        set({
          mockedUserId,
          deploymentCredentials: $credentialsQuery.data,
          projectPermissions: undefined,
          isLoading: $credentialsQuery.isLoading,
        });
        return;
      }

      const projectQuery = createAdminServiceGetProjectWithBearerToken(
        org,
        project,
        accessToken,
        undefined,
        {
          query: {
            enabled: true,
          },
        },
      );

      const unsubProject = projectQuery.subscribe(($projectQuery) => {
        set({
          mockedUserId,
          deploymentCredentials: $credentialsQuery.data,
          projectPermissions: $projectQuery.data?.projectPermissions,
          isLoading: $credentialsQuery.isLoading || $projectQuery.isLoading,
        });
      });

      return unsubProject;
    });

    return unsubCredentials;
  });
}

/**
 * Computes effective project permissions by merging actual permissions with
 * mocked permissions when View As is active.
 *
 * When View As is active and mocked permissions are available, returns the
 * mocked permissions. Otherwise, returns the actual user's permissions.
 */
export function getEffectivePermissions(
  viewAsState: ViewAsState,
  actualPermissions: V1ProjectPermissions | undefined,
): V1ProjectPermissions | undefined {
  return viewAsState.mockedUserId && viewAsState.projectPermissions
    ? viewAsState.projectPermissions
    : actualPermissions;
}
