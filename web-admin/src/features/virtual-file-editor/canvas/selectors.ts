import { derived } from "svelte/store";
import { createAdminServiceListPersonalFiles } from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export function getPersonalCanvases(
  client: RuntimeClient,
  org: string,
  project: string,
) {
  return derived(
    [
      createAdminServiceListPersonalFiles(org, project),
      createRuntimeServiceListResources(client, {}),
    ],
    ([personalFilesResp, resourcesResp]) => {
      const isPending = personalFilesResp.isPending || resourcesResp.isPending;
      const error = resourcesResp.error ?? personalFilesResp.error;

      const personalFileNames = new Set(personalFilesResp.data?.files ?? []);
      const personalCanvases = resourcesResp.data?.resources.filter(
        (r) =>
          r.meta?.name?.kind === ResourceKind.Canvas &&
          personalFileNames.has(r.meta?.name?.name),
      );

      return {
        isPending,
        error,
        data: personalCanvases,
      };
    },
  );
}
