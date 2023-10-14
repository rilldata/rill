<script lang="ts">
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import {
    capitalizeFirstLetter,
    formatDateToCustomString,
  } from "./tableUtils";

  export let id: string;
  export let reportName: string;
  export let lastRun: string;
  export let frequency: string;
  export let author: string;
  export let status: string;
</script>

<div class="flex flex-col gap-y-0.5">
  <div class="flex gap-x-2 items-center">
    <a
      href={`reports/${id}`}
      class="text-gray-700 text-sm font-semibold hover:text-blue-600"
    >
      {reportName}
    </a>
    {#if status === "success"}
      <CheckCircle className="text-blue-500" />
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    <span>Last run {formatDateToCustomString(new Date(lastRun))}</span>
    <span>•</span>
    <span>{capitalizeFirstLetter(frequency)}</span>
    <span>•</span>
    <span>Created by {author}</span>
  </div>
</div>
