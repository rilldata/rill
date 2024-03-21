import { newFileArtifactStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store-new";
import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";

export enum ResourceStatus {
  Idle,
  Busy,
  Errored,
}

export type ResourceStatusState = {
  status: ResourceStatus;
  error?: ErrorType<unknown>;
  resource?: V1Resource;
};

/**
 * Used while saving to wait until either a resource is created or parse has errored.
 */
export function resourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  const lastUpdatedOn = newFileArtifactStore.getLastStateUpdatedOn(filePath);
  return derived(
    [
      newFileArtifactStore.getResourceForFile(
        queryClient,
        instanceId,
        filePath,
      ),
      newFileArtifactStore.getAllErrorsForFile(
        queryClient,
        instanceId,
        filePath,
      ),
    ],
    ([res, errors]) => {
      if (res.isFetching) return { status: ResourceStatus.Busy };
      if (
        (res.isError && (res.error as any).response.status !== 404) ||
        errors.length > 0
      )
        return { status: ResourceStatus.Errored, changed: false };

      if (
        res.data?.meta?.reconcileStatus !==
        V1ReconcileStatus.RECONCILE_STATUS_IDLE
      )
        return { status: ResourceStatus.Busy };

      const changed =
        !lastUpdatedOn ||
        (res.data?.meta?.stateUpdatedOn !== undefined
          ? res.data?.meta?.stateUpdatedOn > lastUpdatedOn
          : false);

      return {
        status: !res.data?.meta?.reconcileError
          ? ResourceStatus.Idle
          : ResourceStatus.Errored,
        changed,
      };
    },
  );
}

export function waitForResourceUpdate(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  return new Promise<boolean>((resolve) => {
    let timer;
    let idled = false;

    const end = (changed: boolean) => {
      unsub?.();
      resolve(changed);
    };

    // eslint-disable-next-line prefer-const
    const unsub = resourceStatusStore(
      queryClient,
      instanceId,
      filePath,
    ).subscribe((status) => {
      if (status.status === ResourceStatus.Busy) return;
      if (timer) clearTimeout(timer);

      const do_end =
        status.status === ResourceStatus.Idle &&
        status.changed !== undefined &&
        status.changed;

      if (idled) {
        end(do_end);
        return;
      } else {
        idled = true;
        timer = setTimeout(() => {
          end(do_end);
        }, 500);
      }
    });
  });
}

/**
 * Assumes the initial resource has been created after a new entity creation.
 */
export function getResourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
  validator?: (res: V1Resource) => boolean,
) {
  return derived(
    [
      newFileArtifactStore.getResourceForFile(
        queryClient,
        instanceId,
        filePath,
      ),
      newFileArtifactStore.getAllErrorsForFile(
        queryClient,
        instanceId,
        filePath,
      ),
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

      let isBusy: boolean;
      if (validator && resourceResp.data) {
        isBusy = !validator(resourceResp.data);
      } else {
        isBusy =
          resourceResp.isFetching ||
          resourceResp.data?.meta?.reconcileStatus !==
            V1ReconcileStatus.RECONCILE_STATUS_IDLE;
      }

      return {
        status: isBusy ? ResourceStatus.Busy : ResourceStatus.Idle,
        resource: resourceResp.data,
      };
    },
  ) as Readable<ResourceStatusState>;
}
