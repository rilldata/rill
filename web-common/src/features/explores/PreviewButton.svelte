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
  export let reconciling: boolean = false;

  const viewDashboard = () => {
    if (!href) return;
    behaviourEvent
      ?.fireNavigationEvent(
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
    type="secondary"
    preload={false}
    compact
    {loading}
    {href}
    {disabled}
    on:click={viewDashboard}
  >
    <div class="flex gap-x-1 items-center">
      <Play size="14px" />
      Preview
    </div>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if reconciling}
      Dashboard preview available after reconciliation
    {:else if disabled}
      File errors must be resolved before previewing
    {:else}
      Preview dashboard
    {/if}
  </TooltipContent>
</Tooltip>
