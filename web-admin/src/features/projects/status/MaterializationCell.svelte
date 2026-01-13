<script lang="ts">
  import { Code2 } from "lucide-svelte";
  import { Database } from "lucide-svelte";

  export let isView: boolean | undefined;
  export let physicalSizeBytes: string | number | undefined;

  $: isLikelyView =
    isView === true ||
    physicalSizeBytes === "-1" ||
    physicalSizeBytes === 0 ||
    !physicalSizeBytes;

  $: label = isLikelyView ? "View" : "Table";
  $: icon = isLikelyView ? Code2 : Database;
  $: color = isLikelyView ? "#0891B2" : "#059669"; // cyan for View, green for Table
</script>

<div class="shrink-0 flex items-center gap-x-1">
  <span
    class="shrink-0 flex items-center gap-x-1 text-[10px] font-medium px-1.5 py-0.5 rounded"
    style="background-color: {color}15; color: {color}"
  >
    <svelte:component this={icon} size="12px" {color} />
    {label}
  </span>
</div>
