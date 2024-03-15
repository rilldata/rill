<script context="module" lang="ts">
  import PivotChip from "./PivotChip.svelte";
  import { type PivotChipData, PivotChipType } from "./types";
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let index: number;
  export let removable: boolean;
  export let hidden: boolean;
</script>

<div
  id={item.id}
  title={item.title}
  data-type={item.type === PivotChipType.Measure ? "measure" : "dimension"}
  data-index={index}
  class="drag-item"
  class:hidden
  class:rounded-full={item.type !== PivotChipType.Measure}
>
  <PivotChip grab {removable} {item} on:mousedown on:remove />
</div>

<style lang="postcss">
  .drag-item {
    @apply w-fit;
  }
</style>
