<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { formatDateToCustomString } from "../tableUtils";

  export let reportTime: string;
  export let errorMessage: string;
</script>

<div class="flex flex-col gap-y-0.5 px-4 py-[5px]">
  <div class="flex gap-x-2 items-center">
    <div class="text-gray-700 text-sm">
      {formatDateToCustomString(new Date(reportTime))}
    </div>
    {#if errorMessage === ""}
      <Tag color="blue">Email sent</Tag>
    {:else}
      <Tag color="red">Failed</Tag>
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    <!-- <span>Export {exportSize}</span> -->
  </div>
</div>
