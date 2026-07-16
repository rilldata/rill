<script lang="ts">
  import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";
  import ChipDragList from "../ChipDragList.svelte";
  import type { FieldType } from "../types";

  export let items: string[] = [];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let onUpdate: (items: string[]) => void;
  export let orientation: "horizontal" | "vertical" = "vertical";
  export let measureFormatting:
    | Record<string, PivotMeasureFormatting>
    | undefined = undefined;
  export let setMeasureFormatting:
    | ((measureName: string, fmt: PivotMeasureFormatting | null) => void)
    | undefined = undefined;
  export let lowerIsBetterMap: Record<string, boolean> = {};
</script>

{#if items?.length > 0}
  <div class="mt-2">
    {#if items?.length > 0}
      <ChipDragList
        {items}
        {displayMap}
        {onUpdate}
        {orientation}
        {measureFormatting}
        {setMeasureFormatting}
        {lowerIsBetterMap}
      />
    {/if}
  </div>
{/if}
