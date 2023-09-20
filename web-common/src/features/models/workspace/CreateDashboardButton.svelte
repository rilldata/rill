<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useModel } from "@rilldata/web-common/features/models/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    createRuntimeServicePutFileAndReconcile,
    V1ReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getName } from "../../entity-management/name-utils";
  import { generateDashboardYAMLForModel } from "../../metrics-views/metrics-internal-store";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: model = $modelQuery.data?.model;
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);

  const queryClient = useQueryClient();
  const createFileMutation = createRuntimeServicePutFileAndReconcile();

  async function handleCreateDashboard() {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    const dashboardYAML = generateDashboardYAMLForModel(
      model as any, // TODO
      newDashboardName
    );

    $createFileMutation.mutate(
      {
        data: {
          instanceId: $runtime.instanceId,
          path: getFilePathFromNameAndType(
            newDashboardName,
            EntityType.MetricsDefinition
          ),
          blob: dashboardYAML,
          create: true,
          createOnly: true,
          strict: false,
        },
      },
      {
        onSuccess: (resp: V1ReconcileResponse) => {
          fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
          goto(`/dashboard/${newDashboardName}`);
          behaviourEvent.fireNavigationEvent(
            newDashboardName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.RightPanel,
            MetricsEventScreenName.Model,
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(
            queryClient,
            $runtime.instanceId,
            resp
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
