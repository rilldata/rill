<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1ReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getName } from "../../entity-management/name-utils";
  import {
    addQuickMetricsToDashboardYAML,
    initBlankDashboardYAML,
  } from "../../metrics-views/metrics-internal-store";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName
  );
  $: model = $getModel.data?.entry?.model;
  $: dashboardNames = useDashboardNames($runtimeStore.instanceId);

  const queryClient = useQueryClient();
  const createFileMutation = useRuntimeServicePutFileAndReconcile();

  async function handleCreateDashboard() {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    const blankDashboardYAML = initBlankDashboardYAML(newDashboardName);
    const fullDashboardYAML = addQuickMetricsToDashboardYAML(
      blankDashboardYAML,
      model
    );

    $createFileMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          path: getFilePathFromNameAndType(
            newDashboardName,
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
          goto(`/dashboard/${newDashboardName}`);
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.RightPanel,
            MetricsEventScreenName.Model,
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(
            queryClient,
            $runtimeStore.instanceId,
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
