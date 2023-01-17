<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    V1ReconcileResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import {
    addQuickMetricsToDashboardYAML,
    initBlankDashboardYAML,
  } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { getName } from "@rilldata/web-local/lib/util/incrementName";
  import { useQueryClient } from "@sveltestack/svelte-query";

  export let modelName: string;
  export let hasError = false;
  export let collapse = false;

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName
  );
  $: model = $getModel.data?.entry?.model;
  $: timestampColumns = selectTimestampColumnFromSchema(model?.schema);
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

<Tooltip alignment="right" distance={16} location="bottom">
  <Button
    disabled={!timestampColumns?.length}
    on:click={handleCreateDashboard}
    type="primary"
  >
    <Add />
    <ResponsiveButtonText {collapse}>Create Dashboard</ResponsiveButtonText>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if hasError}
      Fix the errors in your model to autogenerate dashboard
    {:else if timestampColumns?.length}
      Generate a dashboard based on your model
    {:else}
      Add a timestamp column to your model in order to generate a dashboard
    {/if}
  </TooltipContent>
</Tooltip>
