<script lang="ts">
  import MeasureFormattingControls from "./MeasureFormattingControls.svelte";
  import type { PivotMeasureFormatting } from "./types";

  type Props = {
    // Measures available to format, in display order.
    measures: { name: string; label: string }[];
    measureFormatting: Record<string, PivotMeasureFormatting> | undefined;
    // Emits the new config for a measure, or null to clear formatting.
    onChange: (measureName: string, fmt: PivotMeasureFormatting | null) => void;
  };

  let { measures, measureFormatting, onChange }: Props = $props();
</script>

<div class="flex flex-col gap-y-3">
  {#if measures.length === 0}
    <span class="text-fg-secondary text-xs">No measures to format.</span>
  {/if}
  {#each measures as measure (measure.name)}
    <div class="flex flex-col gap-y-1.5">
      <span class="text-xs font-medium truncate" title={measure.label}>
        {measure.label}
      </span>
      <div class="pl-1">
        <MeasureFormattingControls
          id={measure.name}
          fmt={measureFormatting?.[measure.name]}
          onChange={(fmt: PivotMeasureFormatting | null) =>
            onChange(measure.name, fmt)}
        />
      </div>
    </div>
  {/each}
</div>
