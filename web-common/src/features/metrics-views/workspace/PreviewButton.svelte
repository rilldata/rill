<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { navigating } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { Play } from "lucide-svelte";

  export let disabled: boolean;
  export let status: string[];
  export let metricViewName: string;

  const viewDashboard = () => {
    behaviourEvent
      .fireNavigationEvent(
        metricViewName,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace,
        MetricsEventScreenName.MetricsDefinition,
        MetricsEventScreenName.Dashboard,
      )
      .catch(console.error);
  };
</script>

<Tooltip alignment="middle" distance={5} location="right">
  <Button
    {disabled}
    label="Preview"
    href={`/dashboard/${metricViewName}`}
    on:click={viewDashboard}
    type="brand"
    loading={!!$navigating}
  >
    <Play size="10px" />
    Preview
  </Button>
  <TooltipContent slot="tooltip-content">
    {#each status as message}
      <div>{message}</div>
    {/each}
  </TooltipContent>
</Tooltip>
