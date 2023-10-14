<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";

  export let organization: string;
  export let project: string;
  export let report: string;

  // TODO: fetch report metadata
  const reportMetadata = {
    name: "report-name",
    author: "John Doe",
    dashboard: "Usage metrics",
    label: "First report",
    description: "Description of the report",
    format: "csv",
    limit: 100,
    frequency: "Daily",
    lastRun: "2021-09-01T00:00:00Z",
    nextRun: "2021-09-01T00:00:00Z",
    isValid: true,
  };
</script>

<div class="flex flex-col gap-y-2 max-w-full 2xl:max-w-[1200px]">
  <div class="flex flex-col gap-y-2">
    <div class="flex gap-x-2 items-center">
      <h1 class="text-gray-800 text-base font-medium leading-none">
        {reportMetadata.label}
      </h1>
      <Tag>Report</Tag>
    </div>
    {#if reportMetadata?.description}
      <p class="text-gray-500 text-sm">{reportMetadata.description}</p>
    {/if}
  </div>

  <!-- three columns of metadata -->
  <div class="grid grid-cols-3 gap-x-20 grid-cols-[auto,auto,1fr]">
    <!-- column 1 -->
    <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
      <!-- dashboard name -->
      <span class="text-gray-800 text-sm font-medium">Dashboard</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.dashboard}
      </span>
      <!-- author -->
      <span class="text-gray-800 text-sm font-medium">Author</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.author}
      </span>
      <!-- recipients -->
      <!-- TODO -->
    </div>

    <!-- column 2 -->
    <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
      <!-- last run -->
      <span class="text-gray-800 text-sm font-medium">Format</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.format.toUpperCase()}
      </span>
      <!-- next run -->
      <span class="text-gray-800 text-sm font-medium">Limit</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.limit}
      </span>
    </div>

    <!-- column 3 -->
    <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
      <!-- last run -->
      <span class="text-gray-800 text-sm font-medium">Frequency</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.frequency}
      </span>
      <!-- next run -->
      <span class="text-gray-800 text-sm font-medium">Next run</span>
      <span class="text-gray-700 text-sm font-normal">
        {reportMetadata.nextRun}
      </span>
    </div>
  </div>
</div>
