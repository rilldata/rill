import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getResourceNameForFile,
  useResourceForFile,
} from "@rilldata/web-common/features/entity-management/resources-store";
import {
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";

export enum ResourceStatus {
  Idle,
  Busy,
  Errored,
}

export type ResourceStatusState = {
  status: ResourceStatus;
  error?: unknown;
  resource?: V1Resource;
};

/**
 * Used during creation to wait until either a resource is created or parse has errored.
 */
export function initialResourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): Readable<ResourceStatus> {
  return derived(
    [
      getResourceNameForFile(filePath),
      useProjectParser(queryClient, instanceId),
    ],
    ([resourceName, projectParserResp]) => {
      if (
        !projectParserResp.data ||
        projectParserResp.data.projectParser.state.parseErrors.filter(
          (s) => s.filePath === filePath
        ).length > 0
      ) {
        return ResourceStatus.Errored;
      }

      return !resourceName ? ResourceStatus.Busy : ResourceStatus.Idle;
    }
  );
}

export function waitForResource(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
) {
  return new Promise<void>((resolve) => {
    const unsub = initialResourceStatusStore(
      queryClient,
      instanceId,
      filePath
    ).subscribe((status) => {
      if (status === ResourceStatus.Busy) return;
      unsub();
      resolve();
    });
  });
}

/**
 * Assumes the initial resource has been created after a new entity creation.
 */
export function getResourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): Readable<ResourceStatusState> {
  return derived(
    [
      useResourceForFile(queryClient, instanceId, filePath),
      useProjectParser(queryClient, instanceId),
    ],
    ([resourceResp, projectParserResp]) => {
      if (projectParserResp.isError) {
        return {
          status: ResourceStatus.Errored,
          error: projectParserResp.error,
        };
      }

      if (resourceResp.isError || projectParserResp.isError) {
        return {
          status: ResourceStatus.Errored,
          error: resourceResp.error ?? projectParserResp.error,
        };
      }

      if (projectParserResp.isFetching || resourceResp.isFetching) {
        return {
          status: ResourceStatus.Busy,
        };
      }

      return {
        status:
          resourceResp.data.meta.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_IDLE
            ? ResourceStatus.Idle
            : ResourceStatus.Busy,
        resource: resourceResp.data,
      };
    }
  );
}
