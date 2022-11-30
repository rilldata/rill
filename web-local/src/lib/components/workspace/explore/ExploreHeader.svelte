<script lang="ts">
  import { goto } from "$app/navigation";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { invalidateMetricsViewData } from "../../../svelte-query/queries/metrics-views/invalidation";
  import { useMetaQuery } from "../../../svelte-query/queries/metrics-views/metadata";
  import { Button } from "../../button";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  // TODO: move this "sync" to a more relevant component
  $: if (
    metricViewName &&
    $metaQuery &&
    metricViewName === $metaQuery.data?.name
  ) {
    if (
      !$metaQuery.data?.measures?.length ||
      !$metaQuery.data?.dimensions?.length
    ) {
      goto(`/dashboard/${metricViewName}/edit`);
    } else if (!$metaQuery.isError && !$metaQuery.isFetching) {
      // FIXME: understand this logic before removing invalidateMetricsViewData
      invalidateMetricsViewData(queryClient, metricViewName);
    }
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
  <div class="flex justify-between w-full pt-3 pl-1 pr-4">
    <!-- title element -->
    <h1 style:line-height="1.1">
      <div class="pl-4 pt-1" style:font-size="24px">
        {metricViewName}
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
    <TimeControls metricsDefId={metricViewName} />
    {#key metricViewName}
      <Filters {metricViewName} />
    {/key}
  </div>
</section>
