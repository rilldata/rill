<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "@rilldata/web-common/runtime-client";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import exportMetrics from "./export-metrics";

  let exportMenuOpen = false;

  const ctx = getStateManagers();

  const exportDash = createQueryServiceExport();
  const handleExportMetrics = async (format: V1ExportFormat) => {
    exportMetrics({
      ctx,
      query: exportDash,
      format,
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
    class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
  >
    Export
    <CaretDownIcon
      className="transition-transform {exportMenuOpen && '-rotate-180'}"
      size="10px"
    />
  </button>
  <Menu
    let:toggleFloatingElement
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
