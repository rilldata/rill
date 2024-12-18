<script lang="ts">
  import { navigating } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
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
    // temporary fix. we should do a proper fix by removing this on rill-dev when navigated to preview
    const [, entityType, entityName] = href.split("/");
    if (entityType === "explore") {
      clearExploreSessionStore(entityName, undefined);
    }
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
