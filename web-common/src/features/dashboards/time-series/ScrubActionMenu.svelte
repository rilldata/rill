<script lang="ts">
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { Bot } from "lucide-svelte";
  import { Interval, DateTime } from "luxon";

  export let subInterval: Interval | null;
  export let timeGrain: V1TimeGrain | undefined;
  export let metricsViewName: string;
  export let measureSelectionEnabled: boolean;
  export let onZoom: () => void;

  function handleZoom(e: MouseEvent) {
    e.stopPropagation();
    e.preventDefault();
    onZoom();
  }

  function handleExplain(e: MouseEvent) {
    e.stopPropagation();
    e.preventDefault();
    measureSelection.startAnomalyExplanationChat(metricsViewName);
  }
</script>

{#if subInterval?.isValid && !subInterval.start?.equals(subInterval.end)}
  <div
    class="absolute left-1/2 -top-2 -translate-x-1/2 z-50 pointer-events-auto"
    role="menu"
  >
    <div
      class="border rounded-md bg-popover text-popover-foreground shadow-md min-w-[160px]"
    >
      <!-- Date range header -->
      <div class="px-2 py-1.5 border-b text-xs font-medium text-fg-muted">
        {#if subInterval?.isValid && timeGrain}
          <RangeDisplay interval={subInterval} {timeGrain} />
        {/if}
      </div>

      <!-- Actions -->
      <div class="p-1">
        <button
          class="w-full flex items-center gap-x-2 px-2 py-1.5 text-xs rounded-sm hover:bg-surface-hover cursor-pointer"
          on:click={handleZoom}
          role="menuitem"
        >
          <span class="text-icon-muted">
            <Zoom size="16px" />
          </span>
          <span class="flex-1 text-left">Zoom</span>
          <span class="text-fg-muted">Z</span>
        </button>

        {#if measureSelectionEnabled}
          <button
            class="w-full flex items-center gap-x-2 px-2 py-1.5 text-xs rounded-sm hover:bg-surface-hover cursor-pointer"
            on:click={handleExplain}
            role="menuitem"
          >
            <Bot size={16} class="text-icon-muted" />
            <span class="flex-1 text-left">Explain</span>
            <span class="text-fg-muted">E</span>
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}
