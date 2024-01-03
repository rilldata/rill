<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "@rilldata/web-common/runtime-client";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
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
  <button
    aria-label="Export model data"
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
    class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
  >
    Export
    <CaretDownIcon
      size="10px"
      className="transition-transform {exportMenuOpen && '-rotate-180'}"
    />
  </button>
  <Menu
    minWidth=""
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    slot="floating-element"
    let:toggleFloatingElement
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
