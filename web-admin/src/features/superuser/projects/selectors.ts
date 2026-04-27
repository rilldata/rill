// All mutations in this file bake in `superuserForceAccess: true`. The
// superuser console routinely operates on projects the caller isn't a member
// of, so every mutation needs the flag. Wrapping `mutateAsync` here means call
// sites just pass the business args and cannot forget.
import {
  createAdminServiceGetProject,
  createAdminServiceHibernateProject,
  createAdminServiceRedeployProject,
  createAdminServiceSearchProjectNames,
  createAdminServiceUpdateProject,
} from "@rilldata/web-admin/client";
import { derived } from "svelte/store";

export function searchProjects(namePattern: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern: `%${namePattern}%`, pageSize: 50 },
    { query: { enabled: namePattern.length >= 3 } },
  );
}

export function getProject(org: string, project: string) {
  return createAdminServiceGetProject(
    org,
    project,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 && project.length > 0 } },
  );
}

export function createUpdateProjectMutation() {
  const mutation = createAdminServiceUpdateProject();
  return derived(mutation, ($m) => ({
    ...$m,
    mutateAsync: (vars: {
      org: string;
      project: string;
      data: Parameters<typeof $m.mutateAsync>[0]["data"];
    }) =>
      $m.mutateAsync({
        org: vars.org,
        project: vars.project,
        data: { ...vars.data, superuserForceAccess: true },
      }),
  }));
}

export function createHibernateProjectMutation() {
  const mutation = createAdminServiceHibernateProject();
  return derived(mutation, ($m) => ({
    ...$m,
    mutateAsync: (vars: { org: string; project: string }) =>
      $m.mutateAsync({
        org: vars.org,
        project: vars.project,
        params: { superuserForceAccess: true },
      }),
  }));
}

export function createRedeployProjectMutation() {
  const mutation = createAdminServiceRedeployProject();
  return derived(mutation, ($m) => ({
    ...$m,
    mutateAsync: (vars: { org: string; project: string }) =>
      $m.mutateAsync({
        org: vars.org,
        project: vars.project,
        params: { superuserForceAccess: true },
      }),
  }));
}
