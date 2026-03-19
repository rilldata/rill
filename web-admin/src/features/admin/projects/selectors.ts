// web-admin/src/features/admin/projects/selectors.ts
import {
  createAdminServiceSearchProjectNames,
  createAdminServiceGetProject,
  createAdminServiceUpdateProject,
  createAdminServiceRedeployProject,
  createAdminServiceHibernateProject,
} from "@rilldata/web-admin/client";

export function searchProjects(namePattern: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern, pageSize: 50 },
    { query: { enabled: namePattern.length >= 2 } },
  );
}

export function getProject(org: string, project: string) {
  return createAdminServiceGetProject(org, project);
}

export function createUpdateProjectMutation() {
  return createAdminServiceUpdateProject();
}

export function createRedeployProjectMutation() {
  return createAdminServiceRedeployProject();
}

export function createHibernateProjectMutation() {
  return createAdminServiceHibernateProject();
}
