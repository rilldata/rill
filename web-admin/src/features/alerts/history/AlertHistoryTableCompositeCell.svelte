<script lang="ts">
  import AlertHistoryStatusChip from "@rilldata/web-admin/features/alerts/history/AlertHistoryStatusChip.svelte";
  import { formatRunDate } from "@rilldata/web-admin/features/scheduled-reports/tableUtils";
  import type {
    V1AlertExecution,
    V1AssertionResult,
  } from "@rilldata/web-common/runtime-client";

  export let alertTime: string;
  export let timeZone: string;
  export let currentExecution: V1AlertExecution | null;
  export let result: V1AssertionResult;
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-gray-700 text-sm flex-shrink-0">
    {currentExecution ? "Checking" : "Checked"}
    {formatRunDate(alertTime, timeZone)}
  </div>
  <AlertHistoryStatusChip {currentExecution} {result} />
</div>
