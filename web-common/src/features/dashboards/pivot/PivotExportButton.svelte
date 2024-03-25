<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { builderActions, getAttrs } from "bits-ui";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "../../../runtime-client";
  import { useDashboard } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import exportPivot from "./pivot-export";

  let active = false;

  const ctx = getStateManagers();
  const { runtime, metricsViewName } = ctx;
  const exportDash = createQueryServiceExport();

  $: metricsView = useDashboard($runtime.instanceId, $metricsViewName);

  async function handleExportPivot(format: V1ExportFormat) {
    await exportPivot({
      ctx,
      query: exportDash,
      format,
      timeDimension: $metricsView.data?.metricsView?.spec
        ?.timeDimension as string,
    });
  }
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
      aria-label="Export pivot"
      {...getAttrs([builder])}
      on:click|preventDefault
      use:builderActions={{ builders: [builder] }}
    >
      Export
      <CaretDownIcon
        className="transition-transform {active && '-rotate-180'}"
        size="10px"
      />
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_CSV")}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_PARQUET")}
    >
      Export as Parquet
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_XLSX")}
    >
      Export as XLSX
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
