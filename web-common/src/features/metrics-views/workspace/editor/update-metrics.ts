import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
import { setLineStatuses } from "@rilldata/web-common/components/editor/line-status";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createPersistentDashboardStore } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export function createUpdateMetricsCallback(
  filePath: string,
  metricsDefName: string,
): (event: CustomEvent) => void {
  const debounce = createDebouncer();

  const fileSaver = createRuntimeServicePutFile();

  async function reconcileNewMetricsContent(blob: string) {
    const instanceId = get(runtime).instanceId;
    await get(fileSaver).mutateAsync({
      instanceId,
      path: filePath,
      data: {
        blob,
        create: false,
      },
    });
    // Remove the explorer entity so that everything is reset to defaults next time user navigates to it
    metricsExplorerStore.remove(metricsDefName);
    // Reset local persisted dashboard state for the metrics view
    createPersistentDashboardStore(metricsDefName).reset();
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
        transaction.annotation(skipDebounceAnnotation) !== undefined,
    );

    // get the annotation.
    const debounceAnnotation = debounceTransaction?.annotation(
      skipDebounceAnnotation,
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
