<!-- @component
Tooltip for a leaderboard item, including the number of rows and the ability to copy the value to the clipboard.
-->
<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import {
    humanizeCount,
    humanizePercent,
  } from "@rilldata/web-common/lib/number-formatting/humanizer";

  export let rowCount: number;
  export let totalRowCount: number;

  export let excluded: boolean;
  export let filtered: boolean;
  export let filterExcludeMode: boolean;

  $: filteredStr = filtered ? "filtered " : "";
  $: percent = humanizePercent(rowCount / totalRowCount);
  $: rowCountFormatted = humanizeCount(rowCount);
  $: totalCountFormatted = humanizeCount(totalRowCount);

  // note that rowCountInfoString can never exceed TOOLTIP_STRING_LIMIT==60
  // because the number formatter will always return a string of length <=6 chars
  $: rowCountInfoString = `${rowCountFormatted} of ${totalCountFormatted} ${filteredStr}rows (${percent})`;
</script>

<TooltipShortcutContainer>
  {rowCountInfoString}
</TooltipShortcutContainer>

<TooltipShortcutContainer>
  {#if filtered}
    <div>
      {excluded ? "Include" : "Exclude"}
      this dimension value
    </div>
  {:else}
    <div class="text-ellipsis overflow-hidden whitespace-nowrap">
      Filter {filterExcludeMode ? "out" : "on"}
      this dimension value
    </div>
  {/if}
  <Shortcut>Click</Shortcut>
</TooltipShortcutContainer>
<TooltipShortcutContainer>
  <div>
    <StackingWord key="shift">Copy</StackingWord>
    this dimension value to clipboard
  </div>
  <Shortcut>
    <span style="font-family: var(--system);">⇧</span> + Click
  </Shortcut>
</TooltipShortcutContainer>
