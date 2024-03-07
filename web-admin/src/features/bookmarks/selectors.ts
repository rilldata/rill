import {
  createAdminServiceGetProject,
  createAdminServiceListBookmarks,
} from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(
    orgName,
    projectName,
    {},
    {
      query: {
        enabled: !!orgName && !!projectName,
        select: (resp) => resp.project?.id,
      },
    },
  );
}

// TODO: use the getBookmarks from PR#4185
export function useHomeBookmark(projectId: string, metricsViewName: string) {
  return createAdminServiceListBookmarks(
    {
      projectId,
      resourceKind: ResourceKind.MetricsView,
      resourceName: metricsViewName,
    },
    {
      query: {
        enabled: !!projectId && !!metricsViewName,
        select: (data) => data.bookmarks?.find((b) => b.default),
      },
    },
  );
}
