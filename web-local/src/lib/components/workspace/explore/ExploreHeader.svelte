<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { useMetaQuery } from "../../../svelte-query/dashboards";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricViewName: string;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

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

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.label;
</script>

<section
  class="w-full flex flex-col"
  id="header"
  style:padding-left="{$navigationVisibilityTween * 24}px"
>
  <!-- top row
    title and call to action
  -->
  <div
    style:height="var(--header-height)"
    class="flex items-center justify-between w-full pl-1 pr-4"
  >
    <!-- title element -->
    <h1 style:line-height="1.1" style:margin-top="-1px">
      <div class="pl-4" style:font-family="InterDisplay" style:font-size="20px">
        {displayName || metricViewName}
      </div>
    </h1>
    <!-- top right CTAs -->
    <div style="flex-shrink: 0;">
      <Tooltip distance={8}>
        <Button on:click={() => viewMetrics(metricViewName)} type="secondary">
          Edit Metrics <MetricsIcon size="16px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          Edit this dashboard's metrics & settings
        </TooltipContent>
      </Tooltip>
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
