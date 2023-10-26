<script lang="ts">
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import { formatDateToCustomString } from "../tableUtils";

  export let id: string;
  export let timestamp: string;
  export let exportSize: string;
  export let status: string;
</script>

<div class="flex flex-col gap-y-0.5 px-4 py-[5px]">
  <div class="flex gap-x-2 items-center">
    <div class="text-gray-700 text-sm">
      {formatDateToCustomString(new Date(timestamp))}
    </div>
    <!-- TODO: replace this with status tag -->
    {#if status === "success"}
      <CheckCircleOutline className="text-blue-500" />
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    <span>Export {exportSize}</span>
  </div>
</div>
