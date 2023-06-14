<script lang="ts">
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import RowsViewer from "../rows-viewer/RowsViewer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  export let metricViewName: string;
  let isOpen = false;
  const toggle = () => {
    isOpen = !isOpen;
  };
</script>

<div>
  <button
    class="w-full bg-gray-100 h-7 text-left px-2 border-t border-t-gray-200 text-xs text-gray-800 flex items-center gap-1"
    on:click={toggle}
  >
    <span class="font-bold">Source Data</span> x of X rows
    {#if isOpen}
      <CaretUpIcon size="14px" />
    {:else}
      <CaretDownIcon size="14px" />
    {/if}
  </button>
  {#if isOpen}
    <RowsViewer {metricViewName} />
  {/if}
</div>
