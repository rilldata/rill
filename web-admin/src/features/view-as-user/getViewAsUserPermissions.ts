import {
  type V1ProjectPermissions,
  type V1OrganizationPermissions,
} from "../../client";

/**
 * Project roles and their corresponding permissions.
 * Based on: https://docs.rilldata.com/guide/administration/users-and-access/roles-permissions
 */
const PROJECT_ROLE_PERMISSIONS: Record<
  string,
  Partial<V1ProjectPermissions>
> = {
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
 * Organization roles and their corresponding permissions.
 * Based on: https://docs.rilldata.com/guide/administration/users-and-access/roles-permissions
 */
const ORG_ROLE_PERMISSIONS: Record<
  string,
  Partial<V1OrganizationPermissions>
> = {
  admin: {
    admin: true,
    readOrg: true,
    manageOrg: true,
    readProjects: true,
    createProjects: true,
    manageProjects: true,
    readOrgMembers: true,
    manageOrgMembers: true,
    manageOrgAdmins: true,
  },
  editor: {
    admin: false,
    readOrg: true,
    manageOrg: false,
    readProjects: true,
    createProjects: true,
    manageProjects: false,
    readOrgMembers: true,
    manageOrgMembers: false,
    manageOrgAdmins: false,
  },
  viewer: {
    admin: false,
    guest: false,
    readOrg: true,
    manageOrg: false,
    readProjects: true,
    createProjects: false,
    manageProjects: false,
    readOrgMembers: false,
    manageOrgMembers: false,
    manageOrgAdmins: false,
  },
};

/**
 * Returns the default (most restrictive) project permissions for unknown roles.
 */
function getDefaultProjectPermissions(): Partial<V1ProjectPermissions> {
  return {
    readProject: true,
    manageProject: false,
    createMagicAuthTokens: false,
    manageProjectMembers: false,
    manageProjectAdmins: false,
  };
}

/**
 * Returns the default (most restrictive) org permissions for unknown roles.
 */
function getDefaultOrgPermissions(): Partial<V1OrganizationPermissions> {
  return {
    admin: false,
    guest: false,
    readOrg: true,
    manageOrg: false,
    readProjects: true,
    createProjects: false,
    manageProjects: false,
    readOrgMembers: false,
    manageOrgMembers: false,
    manageOrgAdmins: false,
  };
}

/**
 * Maps a role name to project permissions.
 */
export function roleToPermissions(
  roleName: string | undefined,
): Partial<V1ProjectPermissions> {
  if (!roleName) return getDefaultProjectPermissions();
  const normalizedRole = roleName.toLowerCase();
  return (
    PROJECT_ROLE_PERMISSIONS[normalizedRole] ?? getDefaultProjectPermissions()
  );
}

/**
 * Maps an org role name to organization permissions.
 */
export function orgRoleToPermissions(
  roleName: string | undefined,
): Partial<V1OrganizationPermissions> {
  if (!roleName) return getDefaultOrgPermissions();
  const normalizedRole = roleName.toLowerCase();
  return ORG_ROLE_PERMISSIONS[normalizedRole] ?? getDefaultOrgPermissions();
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

/**
 * Computes the effective organization permissions when View As is active.
 * Returns the impersonated user's permissions based on their org role,
 * or the actual permissions if View As is not active.
 */
export function getEffectiveOrgPermissions(
  actualPermissions: V1OrganizationPermissions,
  viewAsOrgRoleName: string | null | undefined,
  isViewAsActive: boolean,
): V1OrganizationPermissions {
  if (!isViewAsActive || !viewAsOrgRoleName) {
    return actualPermissions;
  }

  const viewAsPermissions = orgRoleToPermissions(viewAsOrgRoleName);

  // Return permissions that are the intersection of actual and view-as permissions
  // This ensures we never grant more permissions than the actual user has
  return {
    admin: actualPermissions.admin && viewAsPermissions.admin,
    guest: actualPermissions.guest && viewAsPermissions.guest,
    readOrg: actualPermissions.readOrg && viewAsPermissions.readOrg,
    manageOrg: actualPermissions.manageOrg && viewAsPermissions.manageOrg,
    readProjects:
      actualPermissions.readProjects && viewAsPermissions.readProjects,
    createProjects:
      actualPermissions.createProjects && viewAsPermissions.createProjects,
    manageProjects:
      actualPermissions.manageProjects && viewAsPermissions.manageProjects,
    readOrgMembers:
      actualPermissions.readOrgMembers && viewAsPermissions.readOrgMembers,
    manageOrgMembers:
      actualPermissions.manageOrgMembers && viewAsPermissions.manageOrgMembers,
    manageOrgAdmins:
      actualPermissions.manageOrgAdmins && viewAsPermissions.manageOrgAdmins,
  };
}
