<script lang="ts">
  import PinOff from "@rilldata/web-common/components/icons/PinOff.svelte";
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import * as BitsTooltip from "@rilldata/web-common/components/tooltip-v2";

  export let pinned: boolean;
  export let onTogglePin: () => void;
</script>

<button
  class="h-full aspect-square flex items-center"
  on:click={onTogglePin}
  aria-label={pinned ? "Unpin filter" : "Pin filter"}
>
  <BitsTooltip.Root portal="body">
    <BitsTooltip.Trigger id="pin-measure-filter-tooltip">
      {#if pinned}
        <PinOff size="15px" />
      {:else}
        <Pin size="15px" />
      {/if}
    </BitsTooltip.Trigger>
    <BitsTooltip.Content
      side="right"
      class="z-[100000] max-w-56 bg-gray-700 dark:bg-gray-900 shadow-md text-surface rounded p-2 pt-1 pb-1"
      sideOffset={8}
    >
      Click to pin or unpin : Keep this filter visible at the top so it can’t be
      removed by other users.
    </BitsTooltip.Content>
  </BitsTooltip.Root>
</button>
