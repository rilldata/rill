<script lang="ts">
  import PinOff from "@rilldata/web-common/components/icons/PinOff.svelte";
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import * as BitsTooltip from "@rilldata/web-common/components/tooltip-v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let pinned: boolean;
  export let onTogglePin: () => void;
</script>

<button
  class="h-full aspect-square flex items-center"
  onclick={onTogglePin}
  aria-label={pinned
    ? m.dashboards_filters_unpin()
    : m.dashboards_filters_pin()}
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
      {m.dashboards_filters_pin_tooltip()}
    </BitsTooltip.Content>
  </BitsTooltip.Root>
</button>
