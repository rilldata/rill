import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
import {
  V1PutFileAndReconcileResponse,
  createRuntimeServicePutFileAndReconcile,
} from "@rilldata/web-common/runtime-client";
import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function createUpdateMetricsCallback(
  queryClient: QueryClient,
  metricsDefName: string
): (event: CustomEvent) => void {
  const debounce = createDebouncer();

  const reconcile = createRuntimeServicePutFileAndReconcile();

  async function reconcileNewMetricsContent(blob: string) {
    const instanceId = get(runtime).instanceId;
    const filePath = getFilePathFromNameAndType(
      metricsDefName,
      EntityType.MetricsDefinition
    );
    const resp = (await get(reconcile).mutateAsync({
      data: {
        instanceId: get(runtime).instanceId,
        path: filePath,
        blob,
        create: false,
      },
    })) as V1PutFileAndReconcileResponse;
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
    invalidateAfterReconcile(queryClient, instanceId, resp);
  }

  return function updateMetrics(event) {
    const { content, viewUpdate } = event.detail;
    // immediately reconcile if the user deletes all the content.
    const immediateReconcileFromContentDeletion = !content?.length;

    // check to see if this transaction has a debounce annotation.
    // This will be dispatched in change transactions with the debounceDocUpdateAnnotation
    // added to it.
    const debounceTransaction = viewUpdate.transactions.find(
      (transaction) =>
        transaction.annotation(skipDebounceAnnotation) !== undefined
    );

    // get the annotation.
    const debounceAnnotation = debounceTransaction?.annotation(
      skipDebounceAnnotation
    );
    // We will skip the debounce if the user deletes all the content or there is a skipDebounceAnnotation.
    // See Placeholder.svelte for usage of this annotation.
    // We otherwise debounce to 200ms to prevent unneeded reconciliation thrashing.
    const debounceMS =
      immediateReconcileFromContentDeletion || debounceAnnotation ? 0 : 200;
    debounce(() => {
      reconcileNewMetricsContent(content);
    }, debounceMS);

    // immediately set the line statuses to be empty if the content is empty.
    if (!content?.length) {
      setLineStatuses([], viewUpdate.view);
    }
  };
}
