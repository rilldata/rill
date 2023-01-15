import { goto } from "$app/navigation";
import { EntityType } from "@rilldata/web-common/lib/entity";
import type {
  V1Model,
  V1ReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import {
  addQuickMetricsToDashboardYAML,
  initBlankDashboardYAML,
} from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
import {
  EntityTypeToScreenMap,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import {
  displayName,
  getName,
} from "@rilldata/web-local/lib/util/incrementName";
import { get } from "svelte/store";

export function createDashboardFromModel(
  model: V1Model,
  dashboardDisplayNames,
  dashboardNames,
  createFileMutation,
  queryClient,
  settledCallback = undefined
) {
  overlay.set({
    title: "Creating a dashboard for " + model.name,
  });
  const dashboardFileName = getName(`${model.name}_dashboard`, dashboardNames);
  const newDashboardName = getName(
    displayName(`${model.name}_dashboard`),
    dashboardDisplayNames,
    true
  );
  const blankDashboardYAML = initBlankDashboardYAML(newDashboardName);
  const fullDashboardYAML = addQuickMetricsToDashboardYAML(
    blankDashboardYAML,
    model
  );
  createFileMutation.mutate(
    {
      data: {
        instanceId: get(runtimeStore).instanceId,
        path: getFilePathFromNameAndType(
          dashboardFileName,
          EntityType.MetricsDefinition
        ),
        blob: fullDashboardYAML,
        create: true,
        createOnly: true,
        strict: false,
      },
    },
    {
      onSuccess: (resp: V1ReconcileResponse) => {
        fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
        goto(`/dashboard/${dashboardFileName}`);
        const previousActiveEntity = get(appStore)?.activeEntity?.type;
        navigationEvent.fireEvent(
          dashboardFileName,
          BehaviourEventMedium.Menu,
          MetricsEventSpace.LeftPanel,
          EntityTypeToScreenMap[previousActiveEntity],
          MetricsEventScreenName.Dashboard
        );
        return invalidateAfterReconcile(
          queryClient,
          get(runtimeStore).instanceId,
          resp
        );
      },
      onError: (err) => {
        console.error(err);
      },
      onSettled: () => {
        overlay.set(null);
        if (settledCallback) settledCallback();
      },
    }
  );
}
