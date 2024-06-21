import { goto } from "$app/navigation";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { get } from "svelte/store";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
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
import { getName } from "../../entity-management/name-utils";
import { featureFlags } from "../../feature-flags";
import OptionToCancelAIGeneration from "./OptionToCancelAIGeneration.svelte";

/**
 * TanStack Query does not support mutation cancellation (at least as of v4).
 * Here, we create our own version of `runtimeServiceGenerateMetricsViewFile` that accepts an
 * AbortSignal, which we can use to cancel the request.
 */
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
 * Wrapper function that takes care of common UI side effects on top of creating a dashboard from a table.
 *
 * This function is to be called from all `Generate dashboard with AI` CTAs *outside* of the Metrics Editor.
 */
export function useCreateDashboardFromTableUIAction(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
  folder: string,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
) {
  const isAiEnabled = get(featureFlags.ai);

  // Return a function that can be called to create a dashboard from a table
  return async () => {
    let isAICancelled = false;
    const abortController = new AbortController();

    overlay.set({
      title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} personalizing your dashboard`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort();
            isAICancelled = true;
          },
        },
      },
    });

    // Get a unique name
    const newDashboardName = getName(
      `${tableName}_dashboard`,
      fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
    );
    const newFilePath = `/${folder}/${newDashboardName}.yaml`;

    try {
      // First, request an AI-generated dashboard
      void runtimeServiceGenerateMetricsViewFileWithSignal(
        instanceId,
        {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
          table: tableName,
          path: newFilePath,
          useAi: isAiEnabled, // AI isn't enabled during e2e tests
        },
        abortController.signal,
      );

      // Poll every second until the AI generation is complete or canceled
      while (!isAICancelled) {
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
          await runtimeServiceGetFile(instanceId, { path: newFilePath });
          // success, AI is done
          break;
        } catch (err) {
          // 404 error, AI is not done
        }
      }

      // If the user canceled the AI request, submit another request with `useAi=false`
      if (isAICancelled) {
        await runtimeServiceGenerateMetricsViewFile(instanceId, {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
          table: tableName,
          path: newFilePath,
          useAi: false,
        });
      }

      // Preview
      const previousScreenName = getScreenNameFromPage();
      await goto(`/files${newFilePath}`);
      void behaviourEvent.fireNavigationEvent(
        newDashboardName,
        behaviourEventMedium,
        metricsEventSpace,
        previousScreenName,
        MetricsEventScreenName.Dashboard,
      );
    } catch (err) {
      eventBus.emit("notification", {
        message: "Failed to create a dashboard for " + tableName,
        detail: err.response?.data?.message ?? err.message,
      });
    }

    // Done, remove the overlay
    overlay.set(null);
  };
}

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from a model.
 *
 * This function is to be called from the `Generate dashboard with AI` CTA *inside* of the Metrics Editor.
 */
export async function createDashboardFromTableInMetricsEditor(
  instanceId: string,
  modelName: string,
  filePath: string,
) {
  const isAiEnabled = get(featureFlags.ai);

  const tableName = modelName;
  let isAICancelled = false;
  const abortController = new AbortController();

  overlay.set({
    title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} personalizing your dashboard`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort();
          isAICancelled = true;
        },
      },
    },
  });

  try {
    // First, request an AI-generated dashboard
    void runtimeServiceGenerateMetricsViewFileWithSignal(
      instanceId,
      {
        table: tableName,
        path: filePath,
        useAi: isAiEnabled, // AI isn't enabled during e2e tests
      },
      abortController.signal,
    );

    // Poll every second until the AI generation is complete or canceled
    while (!isAICancelled) {
      await new Promise((resolve) => setTimeout(resolve, 1000));

      try {
        const file = await runtimeServiceGetFile(instanceId, {
          path: filePath,
        });
        if (file.blob !== "") {
          // success, AI is done
          break;
        }
      } catch (err) {
        // 404 error, AI is not done
      }
    }

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (isAICancelled) {
      await runtimeServiceGenerateMetricsViewFile(instanceId, {
        table: tableName,
        path: filePath,
        useAi: false,
      });
    }
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create a dashboard for " + tableName,
      detail: err.response?.data?.message ?? err.message,
    });
  }

  // Done, remove the overlay
  overlay.set(null);
}
