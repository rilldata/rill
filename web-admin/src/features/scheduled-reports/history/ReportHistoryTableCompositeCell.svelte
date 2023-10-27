<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { formatDateToCustomString } from "../tableUtils";

  export let reportTime: string;
  export let errorMessage: string;
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-gray-700 text-sm">
    {formatDateToCustomString(new Date(reportTime))}
  </div>
  {#if errorMessage === ""}
    <Tag color="blue">Email sent</Tag>
  {:else}
    <Tag color="red">Failed</Tag>
  {/if}
</div>
