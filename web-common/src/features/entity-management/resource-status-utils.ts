import {
  ResourceKind,
  useProjectParser,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getAllErrorsForFile,
  getLastStateUpdatedOnByKindAndName,
  getResourceNameForFile,
  useResourceForFile,
} from "@rilldata/web-common/features/entity-management/resources-store";
import {
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, Readable, Unsubscriber } from "svelte/store";

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
 * Used during creation to wait until either a resource is created or parse has errored.
 */
export function initialResourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
): Readable<ResourceStatus> {
  return derived(
    [
      getResourceNameForFile(filePath),
      useProjectParser(queryClient, instanceId),
    ],
    ([resourceName, projectParserResp]) => {
      if (
        !projectParserResp.data ||
        (projectParserResp?.data?.projectParser?.state?.parseErrors?.filter(
          (s) => s.filePath === filePath,
        ).length ?? 0) > 0
      ) {
        return ResourceStatus.Errored;
      }

      return !resourceName ? ResourceStatus.Busy : ResourceStatus.Idle;
    },
  );
}

export function waitForResource(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  return new Promise<void>((resolve) => {
    let unsub: Unsubscriber;
    // eslint-disable-next-line prefer-const
    unsub = initialResourceStatusStore(
      queryClient,
      instanceId,
      filePath,
    ).subscribe((status) => {
      if (status === ResourceStatus.Busy) return;
      unsub?.();
      resolve();
    });
  });
}

/**
 * Used while saving to wait until either a resource is created or parse has errored.
 */
export function resourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
  kind: ResourceKind,
  name: string,
) {
  const lastUpdatedOn = getLastStateUpdatedOnByKindAndName(kind, name);
  return derived(
    [
      useResourceForFile(queryClient, instanceId, filePath),
      getAllErrorsForFile(queryClient, instanceId, filePath),
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
  kind: ResourceKind,
  name: string,
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
      kind,
      name,
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
  );
}
