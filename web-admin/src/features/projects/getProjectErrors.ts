import { MainResourceKinds } from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceListResources,
  V1ParseError,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";

export function getProjectErrors(
  queryClient: QueryClient,
  instanceId: string,
): Readable<Array<V1ParseError>> {
  return derived(
    [
      useProjectParser(queryClient, instanceId),
      createRuntimeServiceListResources(instanceId),
    ],
    ([projectParserResp, resourcesResp]) => {
      const projectParserErrors =
        projectParserResp.data?.projectParser?.state?.parseErrors ?? [];
      const resourceErrors: Array<V1ParseError> =
        resourcesResp.data?.resources
          ?.filter(
            (r) =>
              !!r.meta.reconcileError && MainResourceKinds[r.meta.name.kind],
          )
          .map((r) => ({
            filePath: r.meta.filePaths[0], // TODO: handle multiple files mapping to same resource
            message: r.meta.reconcileError,
          })) ?? [];
      return [...projectParserErrors, ...resourceErrors];
    },
  );
}
