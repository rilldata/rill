<script lang="ts">
  import { goto } from "$app/navigation";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import Calendar from "../../../components/icons/Calendar.svelte";
  import Shortcut from "../../../components/tooltip/Shortcut.svelte";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { useModelTimestampColumns } from "../../models/selectors";

  export let metricViewName: string;
  export let modelName: string;

  let timestampColumns: Array<string>;
  const timestampColumnsQuery = useModelTimestampColumns(
    $runtimeStore.instanceId,
    modelName
  );
  $: timestampColumns = $timestampColumnsQuery?.data;

  $: redirectToScreen = timestampColumns?.length > 0 ? "metrics" : "model";

  function noTimeseriesCTA() {
    if (timestampColumns?.length) {
      goto(`/dashboard/${metricViewName}/edit`);
    } else {
      goto(`/model/${modelName}`);
    }
  }
</script>

<Tooltip location="bottom" distance={8}>
  <div
    on:click={() => noTimeseriesCTA()}
    class="px-3 py-2 flex flex-row items-center gap-x-3 cursor-pointer"
  >
    <span class="ui-copy-icon"><Calendar size="16px" /></span>
    <span class="ui-copy-disabled">No time dimension specified</span>
  </div>
  <TooltipContent slot="tooltip-content" maxWidth="250px">
    Add a time dimension to your {redirectToScreen} to enable time series plots.
    <TooltipShortcutContainer>
      <div class="capitalize">Edit {redirectToScreen}</div>
      <Shortcut>Click</Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
