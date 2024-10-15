<script lang="ts">
  import { createTimeRangeSummary } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let ctx: StateManagers;
  const validSpecQuery = ctx.validSpecStore;
  const timeRangeSummaryQuery = createTimeRangeSummary(ctx);
</script>

{!!$validSpecQuery.data}
{!!$timeRangeSummaryQuery.data}
