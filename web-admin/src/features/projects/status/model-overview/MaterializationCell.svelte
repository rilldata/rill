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
</script>

<div class="shrink-0 flex items-center gap-x-1">
  <span
    class="shrink-0 flex items-center gap-x-1 text-[10px] font-medium px-1.5 py-0.5 rounded
      {isLikelyView
      ? 'bg-cyan-600/15 text-cyan-600'
      : 'bg-emerald-600/15 text-emerald-600'}"
  >
    <!-- Icon inherits text color via currentColor -->
    <svelte:component this={icon} size="12px" />
    {label}
  </span>
</div>
