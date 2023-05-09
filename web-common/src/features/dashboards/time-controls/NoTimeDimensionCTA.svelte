<script lang="ts">
  import { goto } from "$app/navigation";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useModelTimestampColumns } from "@rilldata/web-common/features/models/selectors";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let metricViewName: string;
  export let modelName: string;

  let timestampColumns: Array<string>;
  const timestampColumnsQuery = useModelTimestampColumns(
    $runtime.instanceId,
    modelName
  );
  $: timestampColumns = $timestampColumnsQuery?.data;
  $: isReadOnlyDashboard = $featureFlags.readOnly === true;

  $: redirectToScreen = timestampColumns?.length > 0 ? "metrics" : "model";

  function noTimeseriesCTA() {
    if (isReadOnlyDashboard) return;
    if (timestampColumns?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    } else {
      goto(`/model/${modelName}`);
    }
  }
</script>

<Tooltip location="bottom" distance={8}>
  <button
    on:click={() => noTimeseriesCTA()}
    class="px-3 py-2 flex flex-row items-center gap-x-3 cursor-pointer"
  >
    <span class="ui-copy-icon"><Calendar size="16px" /></span>
    <span class="ui-copy-disabled">No time dimension specified</span>
  </button>
  <TooltipContent slot="tooltip-content" maxWidth="250px">
    {#if isReadOnlyDashboard}
      No time dimension available for this dashboard.
    {:else}
      Add a time dimension to your {redirectToScreen} to enable time series plots.
      <TooltipShortcutContainer>
        <div class="capitalize">Edit {redirectToScreen}</div>
        <Shortcut>Click</Shortcut>
      </TooltipShortcutContainer>
    {/if}
  </TooltipContent>
</Tooltip>
