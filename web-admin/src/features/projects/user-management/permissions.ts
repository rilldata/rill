import type { V1ProjectPermissions } from "@rilldata/web-admin/client";

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permission
 *
 * Checks if a user has admin permissions for a project
 */
export function isProjectAdmin(permissions?: V1ProjectPermissions): boolean {
  return !!permissions?.admin;
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permission
 *
 * Checks if a user has editor permissions for a project
 */
export function isProjectEditor(permissions?: V1ProjectPermissions): boolean {
  return !!(
    permissions?.readProject &&
    permissions?.readProd &&
    permissions?.readProdStatus &&
    permissions?.readProjectMembers &&
    permissions?.manageProjectMembers &&
    permissions?.createMagicAuthTokens &&
    permissions?.manageMagicAuthTokens &&
    permissions?.createReports &&
    permissions?.createAlerts &&
    permissions?.createBookmarks
  );
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permission
 *
 * Checks if a user has viewer permissions for a project
 */
export function isProjectViewer(permissions?: V1ProjectPermissions): boolean {
  return !!(
    permissions?.readProject &&
    permissions?.readProd &&
    permissions?.createReports &&
    permissions?.createAlerts &&
    permissions?.createBookmarks
  );
}
