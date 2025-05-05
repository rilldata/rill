import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has admin permissions for an organization
 */
export function isOrgAdmin(permissions?: V1OrganizationPermissions): boolean {
  return !!permissions?.admin;
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has editor permissions for an organization
 */
export function isOrgEditor(permissions?: V1OrganizationPermissions): boolean {
  return !!(
    permissions?.readOrg &&
    permissions?.readProjects &&
    permissions?.createProjects &&
    permissions?.readOrgMembers &&
    permissions?.manageOrgMembers
  );
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has viewer permissions for an organization
 */
export function isOrgViewer(permissions?: V1OrganizationPermissions): boolean {
  return !!(permissions?.readOrg && permissions?.readProjects);
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has guest permissions for an organization
 */
export function isOrgGuest(permissions?: V1OrganizationPermissions): boolean {
  return !!permissions?.guest;
}
