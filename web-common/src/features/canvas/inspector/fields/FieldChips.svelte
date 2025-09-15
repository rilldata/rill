<script lang="ts">
  import ChipDragList from "../ChipDragList.svelte";
  import type { FieldType } from "../types";

  export let items: string[] = [];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let onUpdate: (items: string[]) => void;
  export let orientation: "horizontal" | "vertical" = "vertical";
</script>

{#if items?.length > 0}
  <div class="mt-2">
    {#if items?.length > 0}
      <ChipDragList {items} {displayMap} {onUpdate} {orientation} />
    {/if}
  </div>
{/if}
