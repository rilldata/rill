import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
import {
  createAdminServiceListOrganizations,
  createAdminServiceListProjectsForOrganization,
  type V1Organization,
} from "../../client";

/**
 * Query selector for organization breadcrumb paths.
 *
 * Uses `select` to transform the raw org list into a PathOption map, and
 * `placeholderData` so the viewingOrg fallback is available immediately
 * (before the query resolves or when the user isn't logged in).
 */
export function useBreadcrumbOrgPaths(
  userLoggedIn: boolean,
  viewingOrg: string | undefined,
  planDisplayName: string | undefined,
) {
  return createAdminServiceListOrganizations(
    { pageSize: 100 },
    {
      query: {
        enabled: userLoggedIn,
        retry: 2,
        refetchOnMount: true,
        placeholderData: {},
        select: (data) =>
          buildOrgPathMap(
            data.organizations ?? [],
            viewingOrg,
            planDisplayName,
          ),
      },
    },
  );
}

/**
 * Query selector for project breadcrumb paths.
 */
export function useBreadcrumbProjectPaths(
  organization: string | undefined,
  readProjects: boolean,
) {
  return createAdminServiceListProjectsForOrganization(
    organization ?? "",
    { pageSize: 100 },
    {
      query: {
        enabled: !!organization && readProjects,
        retry: 2,
        refetchOnMount: true,
        placeholderData: {},
        select: (data) => buildProjectPathMap(data.projects ?? []),
      },
    },
  );
}

/**
 * Builds a PathOption map for the organization breadcrumb segment.
 *
 * The viewingOrg fallback ensures the active org always appears in the
 * breadcrumb, even when the list-orgs response hasn't loaded yet or
 * doesn't include it (e.g. the user has direct project access but
 * isn't a member of the org).
 */
function buildOrgPathMap(
  organizations: V1Organization[],
  viewingOrg: string | undefined,
  planDisplayName: string | undefined,
): Map<string, PathOption> {
  const pathMap = new Map<string, PathOption>();

  organizations.forEach(({ name, displayName }) => {
    if (!name) return;
    pathMap.set(name.toLowerCase(), {
      label: displayName || name,
      pill: planDisplayName,
    });
  });

  if (!viewingOrg) return pathMap;

  if (!pathMap.has(viewingOrg.toLowerCase())) {
    pathMap.set(viewingOrg.toLowerCase(), {
      label: viewingOrg,
      pill: planDisplayName,
    });
  }

  return pathMap;
}

function buildProjectPathMap(
  projects: { name?: string }[],
): Map<string, PathOption> {
  return projects.reduce((map, { name }) => {
    if (!name) return map;
    return map.set(name.toLowerCase(), { label: name, preloadData: false });
  }, new Map<string, PathOption>());
}
