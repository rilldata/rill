import { derived } from "svelte/store";
import {
  createAdminServiceGetProjectMemberUser,
  type V1ProjectPermissions,
} from "../../client";
import { viewAsUserStore, viewAsUserStateStore$ } from "./viewAsUserStore";

/**
 * Project roles and their corresponding permissions.
 * Based on: https://docs.rilldata.com/guide/administration/users-and-access/roles-permissions
 */
const ROLE_PERMISSIONS: Record<string, Partial<V1ProjectPermissions>> = {
  admin: {
    readProject: true,
    manageProject: true,
    createMagicAuthTokens: true,
    manageProjectMembers: true,
    manageProjectAdmins: true,
  },
  editor: {
    readProject: true,
    manageProject: false,
    createMagicAuthTokens: true,
    manageProjectMembers: true,
    manageProjectAdmins: false,
  },
  viewer: {
    readProject: true,
    manageProject: false,
    createMagicAuthTokens: false,
    manageProjectMembers: false,
    manageProjectAdmins: false,
  },
};

/**
 * Returns the default (most restrictive) permissions for unknown roles.
 */
function getDefaultPermissions(): Partial<V1ProjectPermissions> {
  return {
    readProject: true,
    manageProject: false,
    createMagicAuthTokens: false,
    manageProjectMembers: false,
    manageProjectAdmins: false,
  };
}

/**
 * Maps a role name to project permissions.
 */
export function roleToPermissions(
  roleName: string | undefined,
): Partial<V1ProjectPermissions> {
  if (!roleName) return getDefaultPermissions();
  const normalizedRole = roleName.toLowerCase();
  return ROLE_PERMISSIONS[normalizedRole] ?? getDefaultPermissions();
}

/**
 * Creates a query to fetch the impersonated user's project membership.
 * Returns their role name which can be mapped to permissions.
 */
export function useViewAsUserRole(organization: string, project: string) {
  return derived(viewAsUserStore, ($viewAsUser, set) => {
    if (!$viewAsUser?.email || !organization || !project) {
      set(null);
      return;
    }

    // Create the query - this will be reactive to viewAsUser changes
    const query = createAdminServiceGetProjectMemberUser(
      organization,
      project,
      $viewAsUser.email,
      {
        query: {
          enabled: !!$viewAsUser?.email && !!organization && !!project,
        },
      },
    );

    // Subscribe to the query and update our derived store
    const unsubscribe = query.subscribe((result) => {
      set(result.data?.member?.roleName ?? null);
    });

    return unsubscribe;
  });
}

/**
 * Computes the effective project permissions when View As is active.
 * Returns the impersonated user's permissions based on their role,
 * or null if View As is not active.
 */
export function getEffectivePermissions(
  actualPermissions: V1ProjectPermissions,
  viewAsRoleName: string | null | undefined,
  isViewAsActive: boolean,
): V1ProjectPermissions {
  if (!isViewAsActive || !viewAsRoleName) {
    return actualPermissions;
  }

  const viewAsPermissions = roleToPermissions(viewAsRoleName);

  // Return permissions that are the intersection of actual and view-as permissions
  // This ensures we never grant more permissions than the actual user has
  return {
    readProject: actualPermissions.readProject && viewAsPermissions.readProject,
    manageProject:
      actualPermissions.manageProject && viewAsPermissions.manageProject,
    createMagicAuthTokens:
      actualPermissions.createMagicAuthTokens &&
      viewAsPermissions.createMagicAuthTokens,
    manageProjectMembers:
      actualPermissions.manageProjectMembers &&
      viewAsPermissions.manageProjectMembers,
    manageProjectAdmins:
      actualPermissions.manageProjectAdmins &&
      viewAsPermissions.manageProjectAdmins,
  };
}
