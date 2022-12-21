<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricViewName: string;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);

  let displayName;
  // TODO: move this "sync" to a more relevant component
  $: if (
    metricViewName &&
    $metaQuery &&
    metricViewName === $metaQuery.data?.name
  ) {
    if (!$metaQuery.data?.measures?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    }
    displayName = $metaQuery.data.label;
    metricsExplorerStore.sync(metricViewName, $metaQuery.data);
  }

  const viewMetrics = (metricViewName: string) => {
    goto(`/dashboard/${metricViewName}/edit`);

    navigationEvent.fireEvent(
      metricViewName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      MetricsEventScreenName.MetricsDefinition
    );
  };
</script>

<section class="w-full flex flex-col" id="header">
  <!-- top row
    title and call to action
  -->
  <div
    style:height="var(--header-height)"
    class="flex items-center justify-between w-full pl-1 pr-4"
  >
    <!-- title element -->
    <h1 style:line-height="1.1">
      <div
        class="pl-4 "
        style:font-family="InterDisplay"
        style:font-size="20px"
      >
        {displayName || metricViewName}
      </div>
    </h1>
    <!-- top right CTAs -->
    <div style="flex-shrink: 0;">
      <Button on:click={() => viewMetrics(metricViewName)} type="secondary">
        Edit Metrics <MetricsIcon size="16px" />
      </Button>
    </div>
  </div>
  <!-- bottom row -->
  <div class="px-2 pt-1">
    <TimeControls {metricViewName} />
    {#key metricViewName}
      <Filters {metricViewName} />
    {/key}
  </div>
</section>
