import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getAllErrorsForFile,
  getResourceNameForFile,
  useResourceForFile,
} from "@rilldata/web-common/features/entity-management/resources-store";
import {
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable, Unsubscriber } from "svelte/store";

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
    let unsub: Unsubscriber;
    // eslint-disable-next-line prefer-const
    unsub = initialResourceStatusStore(
      queryClient,
      instanceId,
      filePath
    ).subscribe((status) => {
      if (status === ResourceStatus.Busy) return;
      unsub?.();
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
      getAllErrorsForFile(queryClient, instanceId, filePath),
      useProjectParser(queryClient, instanceId),
    ],
    ([resourceResp, errors, projectParserResp]) => {
      if (projectParserResp.isError) {
        return {
          status: ResourceStatus.Errored,
          error: projectParserResp.error,
        };
      }

      if (
        errors.length ||
        (resourceResp.isError && !resourceResp.isFetching) ||
        projectParserResp.isError
      ) {
        return {
          status: ResourceStatus.Errored,
          error: resourceResp.error ?? projectParserResp.error,
        };
      }

      const isBusy =
        resourceResp.isFetching ||
        resourceResp.data?.meta?.reconcileStatus !==
          V1ReconcileStatus.RECONCILE_STATUS_IDLE;

      return {
        status: isBusy ? ResourceStatus.Busy : ResourceStatus.Idle,
        resource: resourceResp.data,
      };
    }
  );
}
