import {
  V1PersonalVirtualFileType,
  createAdminServiceListPersonalVirtualFiles,
  type V1PersonalVirtualFileSummary,
} from "@rilldata/web-admin/client";
import type { CreateQueryResult } from "@tanstack/svelte-query";

export interface PersonalCanvasListResult {
  files: V1PersonalVirtualFileSummary[];
}

/**
 * usePersonalCanvases lists the calling user's personal canvases for a project.
 * Returns an empty list when the project has not enabled the feature or the user has no canvases.
 */
export function usePersonalCanvases(
  org: string,
  project: string,
): CreateQueryResult<PersonalCanvasListResult> {
  return createAdminServiceListPersonalVirtualFiles(
    org,
    project,
    {
      type: V1PersonalVirtualFileType.PERSONAL_VIRTUAL_FILE_TYPE_CANVAS,
    },
    {
      query: {
        enabled: !!org && !!project,
        select: (data) => ({ files: data.files ?? [] }),
      },
    },
  );
}
