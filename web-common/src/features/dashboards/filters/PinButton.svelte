<script lang="ts">
  import PinOff from "@rilldata/web-common/components/icons/PinOff.svelte";
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import * as BitsTooltip from "@rilldata/web-common/components/tooltip-v2";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let pinned: boolean;
  export let onTogglePin: () => void;
</script>

<button
  class="h-full aspect-square flex items-center"
  onclick={onTogglePin}
  aria-label={pinned ? m.filter_unpin() : m.filter_pin()}
>
  <BitsTooltip.Root>
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
      {m.filter_pin_tooltip()}
    </BitsTooltip.Content>
  </BitsTooltip.Root>
</button>
