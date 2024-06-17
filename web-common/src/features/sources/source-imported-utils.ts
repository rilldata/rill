import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export async function checkSourceImported(
  queryClient: QueryClient,
  filePath: string,
) {
  const lastUpdatedOn =
    fileArtifacts.getFileArtifact(filePath).lastStateUpdatedOn;
  if (lastUpdatedOn) return; // For now only show for fresh sources

  waitForResourceUpdate(queryClient, get(runtime).instanceId, filePath)
    .then((success) => {
      if (!success) return;
      sourceImportedPath.set(filePath);
    })
    .catch(console.error);
}

function waitForResourceUpdate(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  return new Promise<boolean>((resolve) => {
    const end = (changed: boolean) => {
      unsub?.();
      resolve(changed);
    };

    // eslint-disable-next-line prefer-const
    const unsub = sourceImportedStore(
      queryClient,
      instanceId,
      filePath,
    ).subscribe(({ done, errored }) => {
      if (!done) return;
      end(!errored);
    });
  });
}

/**
 * Used while saving to wait until either a resource is created or parse has errored.
 */
function sourceImportedStore(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  const artifact = fileArtifacts.getFileArtifact(filePath);
  return derived(
    [
      artifact.getResource(queryClient, instanceId),
      artifact.getAllErrors(queryClient, instanceId),
    ],
    ([res, errors]) => {
      if (res.isFetching) return { done: false, errored: false };
      if (
        (res.isError && (res.error as any).response.status !== 404) ||
        errors.length > 0
      )
        return { done: true, errored: true };

      if (
        res.data?.meta?.reconcileStatus !==
        V1ReconcileStatus.RECONCILE_STATUS_IDLE
      )
        return { done: false, errored: false };

      return {
        done: !!res.data?.source?.state?.table,
        errored: false,
      };
    },
  );
}
