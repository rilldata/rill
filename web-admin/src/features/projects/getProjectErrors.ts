import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceListResources,
  V1ParseError,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";

export function getProjectErrors(
  queryClient: QueryClient,
  instanceId: string
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
          ?.filter((r) => !!r.meta.reconcileError)
          .map((r) => ({
            filePath: r.meta.filePaths[0],
            message: r.meta.reconcileError,
          })) ?? [];
      return [...projectParserErrors, ...resourceErrors];
    }
  );
}
