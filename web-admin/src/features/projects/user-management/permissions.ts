import type { V1ProjectPermissions } from "@rilldata/web-admin/client";

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permissions
 *
 * Checks if a user has admin permissions for a project
 */
export function isProjectAdmin(permissions?: V1ProjectPermissions): boolean {
  return !!permissions?.admin;
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permissions
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
    // FIXME: when user is editor, isProjectEditor returns false
    // https://www.notion.so/User-Management-Role-Based-Access-Control-RBAC-Enhancements-8d331b29d9b64d87bca066e06ef87f54?d=1dfba33c8f57807f93d0001cb3b0f23d&pvs=4#1dfba33c8f578065bd26cd44df09e830
    // permissions?.createMagicAuthTokens &&
    // permissions?.manageMagicAuthTokens &&
    permissions?.createReports &&
    permissions?.createAlerts &&
    permissions?.createBookmarks
  );
}

/**
 * @source https://docs.rilldata.com/manage/roles-permissions#project-level-permissions
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
