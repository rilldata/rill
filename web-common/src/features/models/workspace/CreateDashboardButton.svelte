<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
  import { useSchemaForTable } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModel } from "@rilldata/web-common/features/models/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getName } from "../../entity-management/name-utils";
  import { generateDashboardYAMLForModel } from "../../metrics-views/metrics-internal-store";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  const queryClient = useQueryClient();

  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);

  $: modelSchema = useSchemaForTable(
    $runtime.instanceId,
    $modelQuery.data?.model
  );

  const createFileMutation = createRuntimeServicePutFile();

  async function handleCreateDashboard() {
    if (!$modelQuery.data?.model) {
      return;
    }

    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    await waitUntil(() => !!$modelSchema.data?.schema);
    const dashboardYAML = generateDashboardYAMLForModel(
      modelName,
      $modelSchema.data?.schema,
      newDashboardName
    );

    $createFileMutation.mutate(
      {
        instanceId: $runtime.instanceId,
        path: getFileAPIPathFromNameAndType(
          newDashboardName,
          EntityType.MetricsDefinition
        ),
        data: {
          blob: dashboardYAML,
          create: true,
          createOnly: true,
        },
      },
      {
        onSuccess: async () => {
          await waitForResource(
            queryClient,
            $runtime.instanceId,
            getFilePathFromNameAndType(
              newDashboardName,
              EntityType.MetricsDefinition
            )
          );
          goto(`/dashboard/${newDashboardName}`);
          behaviourEvent.fireNavigationEvent(
            newDashboardName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.RightPanel,
            MetricsEventScreenName.Model,
            MetricsEventScreenName.Dashboard
          );
        },
        onError: (err) => {
          console.error(err);
        },
        onSettled: () => {
          overlay.set(null);
        },
      }
    );
  }
</script>

<Tooltip alignment="right" distance={8} location="bottom">
  <Button on:click={handleCreateDashboard} type="primary">
    <IconSpaceFixer pullLeft pullRight={collapse}>
      <Add />
    </IconSpaceFixer>
    <ResponsiveButtonText {collapse}>Create Dashboard</ResponsiveButtonText>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model to autogenerate dashboard
    {:else}
      Create a dashboard from this model
    {/if}
  </TooltipContent>
</Tooltip>
