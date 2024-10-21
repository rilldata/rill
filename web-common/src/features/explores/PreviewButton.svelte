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
  import Play from "svelte-radix/Play.svelte";

  export let disabled: boolean;
  export let href: string | null;

  const viewDashboard = () => {
    if (!href) return;
    behaviourEvent
      .fireNavigationEvent(
        href,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace,
        MetricsEventScreenName.MetricsDefinition,
        MetricsEventScreenName.Dashboard,
      )
      .catch(console.error);
  };

  $: loading = $navigating?.to?.url.pathname === href;
</script>

<Tooltip distance={8} location="left">
  <Button
    label="Preview"
    square
    {loading}
    on:click={viewDashboard}
    type="secondary"
    {href}
    {disabled}
  >
    <Play size="16px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if disabled}
      File errors must be resolved before previewing
    {:else}
      Preview dashboard
    {/if}
  </TooltipContent>
</Tooltip>
