<script lang="ts">
  import AlertHistoryStatusChip from "@rilldata/web-admin/features/alerts/history/AlertHistoryStatusChip.svelte";
  import { formatRunDate } from "@rilldata/web-admin/features/scheduled-reports/tableUtils";
  import {
    V1AssertionStatus,
    type V1AlertExecution,
    type V1AssertionResult,
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
  {#if result.status === V1AssertionStatus.ASSERTION_STATUS_ERROR}
    <span class="text-red-600">{result.errorMessage}</span>
  {/if}
</div>
