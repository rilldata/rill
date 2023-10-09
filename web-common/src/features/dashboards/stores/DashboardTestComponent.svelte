<script lang="ts">
  import {
    createTimeRangeSummary,
    useMetaQuery,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let ctx: StateManagers;
  const metaQuery = useMetaQuery(ctx);
  const timeRangeSummaryQuery = createTimeRangeSummary(ctx);
</script>

{!!$metaQuery.data}
{!!$timeRangeSummaryQuery.data}
