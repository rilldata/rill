<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { generateMeasuresAndDimension } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import Explore from "@rilldata/web-local/lib/components/icons/Explore.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/redux-store/source/source-selectors";

  export let modelName: string;
  export let hasError = false;
  export let width = undefined;

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName
  );
  $: model = $getModel.data?.entry?.model;
  $: timestampColumns = selectTimestampColumnFromSchema(model?.schema);

  const metricMigrate = useRuntimeServicePutFileAndReconcile();

  async function handleCreateMetric() {
    const metricsLabel = `${model?.name}_dashboard`;
    const generatedYAML = generateMeasuresAndDimension(model, {
      display_name: metricsLabel,
    });

    await $metricMigrate.mutateAsync({
      data: {
        instanceId: $runtimeStore.instanceId,
        path: `dashboards/${metricsLabel}.yaml`,
        blob: generatedYAML,
        create: true,
      },
    });

    navigationEvent.fireEvent(
      metricsLabel,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Model,
      MetricsEventScreenName.Dashboard
    );

    goto(`/dashboard/${metricsLabel}/edit`);
  }
</script>

<Tooltip alignment="right" distance={16} location="bottom">
  <Button
    disabled={!timestampColumns?.length}
    on:click={handleCreateMetric}
    type="primary"
  >
    <ResponsiveButtonText {width}>Create Dashboard</ResponsiveButtonText>
    <Explore size="16px" /></Button
  >
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
