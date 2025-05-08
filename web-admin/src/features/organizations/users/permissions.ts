import type {
  V1OrganizationPermissions,
  V1OrganizationRole,
} from "@rilldata/web-admin/client";

/**
 * Gets the organization role from the list of roles based on the permissions
 */
export function getOrgRole(
  permissions?: V1OrganizationPermissions,
  roles?: V1OrganizationRole[],
): V1OrganizationRole | undefined {
  if (!permissions || !roles) return undefined;

  // Find the role that matches the user's permissions
  return roles.find((role) => {
    const rolePerms = role.permissions;
    if (!rolePerms) return false;

    // Check if all permissions in the role match the user's permissions
    return Object.entries(rolePerms).every(
      ([key, value]) =>
        permissions[key as keyof V1OrganizationPermissions] === value,
    );
  });
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has admin permissions for an organization
 */
export function isOrgAdmin(
  permissions?: V1OrganizationPermissions,
  roles?: V1OrganizationRole[],
): boolean {
  const role = getOrgRole(permissions, roles);
  return role?.name === "admin";
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has editor permissions for an organization
 */
export function isOrgEditor(
  permissions?: V1OrganizationPermissions,
  roles?: V1OrganizationRole[],
): boolean {
  const role = getOrgRole(permissions, roles);
  return role?.name === "editor";
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has viewer permissions for an organization
 */
export function isOrgViewer(
  permissions?: V1OrganizationPermissions,
  roles?: V1OrganizationRole[],
): boolean {
  const role = getOrgRole(permissions, roles);
  return role?.name === "viewer";
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions
 *
 * Checks if a user has guest permissions for an organization
 */
export function isOrgGuest(
  permissions?: V1OrganizationPermissions,
  roles?: V1OrganizationRole[],
): boolean {
  const role = getOrgRole(permissions, roles);
  return role?.name === "guest";
}
