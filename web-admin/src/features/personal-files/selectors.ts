import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";
import { createAdminServiceListPersonalFiles } from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export function getPersonalFilteredResources(
  client: RuntimeClient,
  org: string,
  project: string,
  kind: ResourceKind,
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
      const personalFiles = resourcesResp.data?.resources?.filter(
        (r) =>
          r.meta?.name?.kind === kind &&
          personalFileNames.has(r.meta?.name?.name),
      );

      return {
        isPending,
        error,
        data: personalFiles,
      };
    },
  );
}

// TODO: maybe move to resources selector?
export function getPersonalFilteredResourceByName(
  client: RuntimeClient,
  name: string,
) {
  return derived(
    createRuntimeServiceListResources(client, {}),
    (resourcesResp) => {
      const resource = resourcesResp.data?.resources?.find(
        (r) => r.meta?.name?.name === name,
      );
      const isPending = resourcesResp.isPending;
      const error = resourcesResp.error;

      return {
        isPending,
        error,
        data: resource,
      };
    },
  );
}
