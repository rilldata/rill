import { createAdminServiceGetProject } from "@rilldata/web-admin/client";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}
