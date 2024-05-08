import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export enum ResourceStatus {
  Idle,
  Busy,
  Errored,
}

/**
 * Used while saving to wait until either a resource is created or parse has errored.
 */
export function resourceStatusStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  const artifact = fileArtifacts.getFileArtifact(filePath);
  const lastUpdatedOn = artifact.lastStateUpdatedOn;
  return derived(
    [
      artifact.getResource(queryClient, instanceId),
      artifact.getAllErrors(queryClient, instanceId),
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

// TODO: have a cleaner method and add to FileArtifact
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
