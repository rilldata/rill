<script lang="ts">
  import { Code2, Database } from "lucide-svelte";
  import { isLikelyView as checkIsLikelyView } from "./utils";

  export let isView: boolean | undefined;
  export let physicalSizeBytes: string | number | undefined;

  $: likelyView = checkIsLikelyView(isView, physicalSizeBytes);
</script>

{#if likelyView !== undefined}
  <div class="shrink-0 flex items-center gap-x-1">
    <span
      class="shrink-0 flex items-center gap-x-1 text-[10px] font-medium px-1.5 py-0.5 rounded
        {likelyView
        ? 'bg-cyan-600/15 text-cyan-600'
        : 'bg-emerald-600/15 text-emerald-600'}"
    >
      <svelte:component this={likelyView ? Code2 : Database} size="12px" />
      {likelyView ? "View" : "Table"}
    </span>
  </div>
{/if}
