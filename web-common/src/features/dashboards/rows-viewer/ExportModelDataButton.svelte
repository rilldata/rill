<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "@rilldata/web-common/runtime-client";
  import exportMetrics from "./export-metrics";

  export let metricViewName;
  let exportMenuOpen = false;

  const timeControlStore = useTimeControlStore(getStateManagers());

  const exportDash = createQueryServiceExport();
  const handleExportMetrics = async (format: V1ExportFormat) => {
    exportMetrics({
      metricViewName,
      query: exportDash,
      format,
      timeControlStore,
    });
  };
</script>

<WithTogglableFloatingElement
  alignment="end"
  distance={8}
  let:toggleFloatingElement
  location="top"
  on:close={() => (exportMenuOpen = false)}
  on:open={() => (exportMenuOpen = true)}
>
  <IconButton
    ariaLabel="Export model data"
    disableTooltip={exportMenuOpen}
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
    ><span class="text-gray-800">
      <Export size="18px" />
    </span>
    <svelte:fragment slot="tooltip-content">Export model data</svelte:fragment>
  </IconButton>
  <Menu
    minWidth=""
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportMetrics("EXPORT_FORMAT_CSV");
      }}
    >
      Export as CSV
    </MenuItem>
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportMetrics("EXPORT_FORMAT_PARQUET");
      }}
    >
      Export as Parquet
    </MenuItem>
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportMetrics("EXPORT_FORMAT_XLSX");
      }}
    >
      Export as XLSX
    </MenuItem>
  </Menu>
</WithTogglableFloatingElement>
