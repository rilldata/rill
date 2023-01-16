<script lang="ts">
  import { goto } from "$app/navigation";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricViewName: string;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);

  let displayName;
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

<section
  class="w-full flex flex-col"
  id="header"
  style:padding-left="{$navigationVisibilityTween * 24}px"
>
  <!-- top row
    title and call to action
  -->

  <!-- bottom row -->
  <div class="px-2 pt-1">
    <TimeControls {metricViewName} />
    {#key metricViewName}
      <Filters {metricViewName} />
    {/key}
  </div>
</section>
