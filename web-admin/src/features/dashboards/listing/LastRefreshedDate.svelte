<script lang="ts">
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useDashboardV2 } from "./selectors";
  import { timeAgo } from "./utils";

  $: dashboardName = $page.params.dashboard;

  $: dashboard = useDashboardV2($runtime?.instanceId, dashboardName);
  $: lastRefreshedDate =
    $dashboard?.data?.refreshedOn && new Date($dashboard.data.refreshedOn);
</script>

{#if lastRefreshedDate}
  <Tooltip distance={8}>
    <div class="text-[11px] text-gray-600">
      Last refreshed {timeAgo(lastRefreshedDate)}
    </div>
    <TooltipContent slot="tooltip-content">
      {lastRefreshedDate.toLocaleString()}
    </TooltipContent>
  </Tooltip>
{/if}
