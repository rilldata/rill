<script lang="ts">
  import { navigating } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
  import { clearMostRecentExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { Play } from "lucide-svelte";

  export let disabled: boolean;
  export let href: string | null;
  export let reconciling: boolean = false;

  const viewDashboard = () => {
    if (!href) return;
    // temporary fix. we should do a proper fix by removing this on rill-dev when navigated to preview
    const [, entityType, entityName] = href.split("/");
    if (entityType === "explore") {
      clearExploreSessionStore(entityName, undefined);
      clearMostRecentExploreState(entityName, undefined);
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
    label={m.explores_preview_button()}
    type="secondary"
    preload={false}
    compact
    {loading}
    {href}
    {disabled}
    onClick={viewDashboard}
  >
    <div class="flex gap-x-1 items-center">
      <Play size={14} />
      {m.explores_preview_button()}
    </div>
  </Button>
  <TooltipContent slot="tooltip-content">
    {#if reconciling}
      {m.explores_preview_tooltip_reconciling()}
    {:else if disabled}
      {m.explores_preview_tooltip_disabled()}
    {:else}
      {m.explores_preview_tooltip_default()}
    {/if}
  </TooltipContent>
</Tooltip>
