<script lang="ts">
  import { formatRunDate } from "@rilldata/web-admin/features/scheduled-reports/tableUtils";
  import type { V1AssertionResult } from "@rilldata/web-common/runtime-client";

  export let alertTime: string;
  export let timeZone: string;
  export let result: V1AssertionResult; // TODO
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-gray-700 text-sm">
    Checking {formatRunDate(alertTime, timeZone)}
  </div>
</div>
