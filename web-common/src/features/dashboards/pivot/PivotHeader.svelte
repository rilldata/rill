<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import DragList from "./DragList.svelte";
  import { getFormattedHeaderValues } from "./pivot-utils";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
    metricsViewName,
  } = stateManagers;

  $: headerData = getFormattedHeaderValues(
    $dashboardStore?.pivot,
    $visibleMeasures,
    $visibleDimensions,
  );
</script>

<div class="header">
  <div class="header-row">
    <span class="row-label"> <Column size="16px" /> Columns</span>
    <DragList
      removable
      items={headerData.columns}
      style="horizontal"
      on:update={(e) => {
        metricsExplorerStore.setPivotColumns(
          $metricsViewName,
          e.detail?.map((item) => item.id),
        );
      }}
    />
  </div>
  <div class="header-row">
    <span class="row-label"> <Row size="16px" /> Rows</span>

    <DragList
      removable
      on:update={(e) => {
        metricsExplorerStore.setPivotRows(
          $metricsViewName,
          e.detail?.map((item) => item.id),
        );
      }}
      items={headerData.rows}
      style="horizontal"
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col;
    border-bottom: 1px solid #ddd;
  }
  .header-row {
    @apply flex items-center gap-x-2 px-2 py-1;
  }
  .row-label {
    @apply flex items-center gap-x-1 w-20;
  }
</style>
