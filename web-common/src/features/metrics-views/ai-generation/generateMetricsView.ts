import { goto } from "$app/navigation";
import { get } from "svelte/store";
import { notifications } from "../../../components/notifications";
import { appScreen } from "../../../layout/app-store";
import { overlay } from "../../../layout/overlay-store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import type { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  type MetricsEventSpace,
} from "../../../metrics/service/MetricsTypes";
import {
  RuntimeServiceGenerateMetricsViewFileBody,
  V1GenerateMetricsViewFileResponse,
  runtimeServiceGenerateMetricsViewFile,
  runtimeServiceGetFile,
} from "../../../runtime-client";
import httpClient from "../../../runtime-client/http-client";
import { useDashboardFileNames } from "../../dashboards/selectors";
import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
import { getName } from "../../entity-management/name-utils";
import { EntityType } from "../../entity-management/types";
import CancelGeneration from "./CancelGeneration.svelte";

const runtimeServiceGenerateMetricsViewFileWithSignal = (
  instanceId: string,
  runtimeServiceGenerateMetricsViewFileBody: RuntimeServiceGenerateMetricsViewFileBody,
  signal: AbortSignal,
) => {
  return httpClient<V1GenerateMetricsViewFileResponse>({
    url: `/v1/instances/${instanceId}/files/generate-metrics-view`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: runtimeServiceGenerateMetricsViewFileBody,
    signal,
  });
};

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from a table.
 */
export function useCreateDashboardFromTableUIAction(
  instanceId: string,
  tableName: string,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
  toggleContextMenu: () => void = () => {},
  goToEditor = false,
) {
  const dashboardNames = useDashboardFileNames(instanceId);

  // abort signal for AI generation
  const abortController = new AbortController();
  let isAICancelled = false;

  // Return a function that can be called to create a dashboard from a table
  return async () => {
    overlay.set({
      title: "Hang tight! AI is personalizing your dashboard",
      component: CancelGeneration,
      componentProps: {
        onCancel: () => {
          abortController.abort();
          isAICancelled = true;
        },
      },
    });

    toggleContextMenu(); // TODO: see if we can bring this out of this function

    const newDashboardName = getName(
      `${tableName}_dashboard`,
      get(dashboardNames).data ?? [],
    );

    try {
      console.log("Using AI to generate dashboard for " + tableName);
      const newFilePath = getFilePathFromNameAndType(
        newDashboardName,
        EntityType.MetricsDefinition,
      );

      void runtimeServiceGenerateMetricsViewFileWithSignal(
        instanceId,
        {
          table: tableName,
          path: newFilePath,
          useAi: true,
        },
        abortController.signal,
      );

      console.log("Waiting for AI...");
      // Poll until the AI generation is complete or canceled
      while (!isAICancelled) {
        // Wait 1 second
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
          await runtimeServiceGetFile(instanceId, newFilePath);
          // AI is done
          break;
        } catch (err) {
          // AI is not done
        }
      }

      // If canceled, then submit another with AI=false
      if (isAICancelled) {
        console.log("AI was canceled");
        await runtimeServiceGenerateMetricsViewFile(instanceId, {
          table: tableName,
          path: newFilePath,
          useAi: false,
        });
      }

      if (goToEditor) {
        await goto(`/dashboard/${newDashboardName}/edit`);
        void behaviourEvent.fireNavigationEvent(
          newDashboardName,
          behaviourEventMedium,
          metricsEventSpace,
          get(appScreen)?.type,
          MetricsEventScreenName.MetricsDefinition,
        );
      } else {
        await goto(`/dashboard/${newDashboardName}`);
        void behaviourEvent.fireNavigationEvent(
          newDashboardName,
          behaviourEventMedium,
          metricsEventSpace,
          get(appScreen)?.type,
          MetricsEventScreenName.Dashboard,
        );
      }
    } catch (err) {
      notifications.send({
        message: "Failed to create a dashboard for " + tableName,
        detail: err.response?.data?.message ?? err.message,
      });
    }
    overlay.set(null);
  };
}
