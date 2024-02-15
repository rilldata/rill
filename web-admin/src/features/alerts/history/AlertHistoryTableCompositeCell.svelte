<script lang="ts">
  import AlertHistoryStatusChip from "@rilldata/web-admin/features/alerts/history/AlertHistoryStatusChip.svelte";
  import { formatRunDate } from "@rilldata/web-admin/features/scheduled-reports/tableUtils";
  import {
    type V1AssertionResult,
    V1AssertionStatus,
  } from "@rilldata/web-common/runtime-client";

  export let alertTime: string;
  export let timeZone: string;
  export let result: V1AssertionResult;
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-gray-700 text-sm">
    {result.status === V1AssertionStatus.ASSERTION_STATUS_UNSPECIFIED
      ? "Checking"
      : "Checked"}
    {formatRunDate(alertTime, timeZone)}
  </div>
  <AlertHistoryStatusChip {result} />
</div>
