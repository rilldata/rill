<script lang="ts">
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
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

<a href={`reports/${id}`} class="flex flex-col gap-y-0.5 group px-4 py-[5px]">
  <div class="flex gap-x-2 items-center">
    <ReportIcon size={"14px"} className="text-slate-500" />
    <div class="text-gray-700 text-sm font-semibold group-hover:text-blue-600">
      {reportName}
    </div>
    {#if status === "success"}
      <CheckCircleOutline className="text-blue-500" />
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    <span>Last run {formatDateToCustomString(new Date(lastRun))}</span>
    <span>•</span>
    <span>{capitalizeFirstLetter(frequency)}</span>
    <span>•</span>
    <span>Created by {author}</span>
  </div>
</a>
