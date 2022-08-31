<script lang="ts">
  import { FileExportType } from "$common/data-modeler-service/ModelActions";
  import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
  import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import {
    ApplicationStore,
    config as appConfig,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import { Button } from "$lib/components/button";
  import WithTogglableFloatingElement from "$lib/components/floating-element/WithTogglableFloatingElement.svelte";
  import Export from "$lib/components/icons/Export.svelte";
  import { Menu, MenuItem } from "$lib/components/menu";

  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "$lib/util/formatters";
  import { getContext } from "svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  import notification from "$lib/components/notifications";
  import PanelCTA from "$lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "$lib/components/panel/ResponsiveButtonText.svelte";
  import WithModelResultTooltip from "../WithModelResultTooltip.svelte";
  export let containerWidth = 0;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  const appStore = getContext("rill:app:store") as ApplicationStore;

  let contextMenuOpen = false;

  const onExport = async (fileType: FileExportType) => {
    let extension = ".csv";
    if (fileType === FileExportType.Parquet) {
      extension = ".parquet";
    }
    const exportFilename = currentModel.name.replace(".sql", extension);

    const exportResp = await dataModelerService.dispatch(fileType, [
      currentModel.id,
      exportFilename,
    ]);

    if (exportResp.status === ActionStatus.Success) {
      window.open(
        `${
          appConfig.server.serverUrl
        }/api/file/export?fileName=${encodeURIComponent(exportFilename)}`
      );
    } else if (exportResp.status === ActionStatus.Failure) {
      notification.send({
        message: `Failed to export.\n${exportResp.messages
          .map((message) => message.message)
          .join("\n")}`,
      });
    }
  };

  let rollup;
  let tables;
  // get source tables?
  let sourceTableReferences;

  /** Select the explicit ID to prevent unneeded reactive updates in currentModel */
  $: activeEntityID = $appStore?.activeEntity?.id;

  let currentModel: PersistentModelEntity;
  $: currentModel =
    activeEntityID && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  let currentDerivedModel: DerivedModelEntity;
  $: currentDerivedModel =
    activeEntityID && $derivedModelStore?.entities
      ? $derivedModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  // get source table references.
  $: if (currentDerivedModel?.sources) {
    sourceTableReferences = currentDerivedModel?.sources;
  }

  // map and filter these source tables.
  $: if (sourceTableReferences?.length) {
    tables = sourceTableReferences
      .map((sourceTableReference) => {
        const table = $persistentTableStore.entities.find(
          (t) => sourceTableReference.name === t.tableName
        );
        if (!table) return undefined;
        return $derivedTableStore.entities.find(
          (derivedTable) => derivedTable.id === table.id
        );
      })
      .filter((t) => !!t);
  } else {
    tables = [];
  }

  $: outputRowCardinalityValue = currentDerivedModel?.cardinality;

  let inputRowCardinalityValue;
  $: if (tables?.length)
    inputRowCardinalityValue = tables.reduce(
      (acc, v) => acc + v.cardinality,
      0
    );

  $: if (
    (inputRowCardinalityValue !== undefined &&
      outputRowCardinalityValue !== undefined) ||
    inputRowCardinalityValue
  ) {
    rollup = outputRowCardinalityValue / inputRowCardinalityValue;
  }

  function validRollup(number) {
    return rollup !== Infinity && rollup !== -Infinity && !isNaN(number);
  }

  // compute column delta
  let inputColumnNum;
  $: if (tables?.length)
    inputColumnNum = tables.reduce(
      (acc, v: DerivedTableEntity) => acc + v.profile.length,
      0
    );
  $: outputColumnNum = currentDerivedModel?.profile?.length;
  $: columnDelta = outputColumnNum - inputColumnNum;

  $: modelHasError = !!currentDerivedModel?.error;
</script>

<PanelCTA side="right" let:width>
  <Tooltip
    location="left"
    alignment="middle"
    distance={16}
    suppress={contextMenuOpen}
  >
    <!-- attach floating element right here-->
    <WithTogglableFloatingElement
      location="left"
      alignment="start"
      distance={16}
      let:toggleFloatingElement
      bind:active={contextMenuOpen}
    >
      <Button
        disabled={modelHasError}
        type="secondary"
        on:click={toggleFloatingElement}
      >
        <ResponsiveButtonText {width}>Export Results</ResponsiveButtonText>
        <Export size="16px" />
      </Button>
      <Menu
        dark
        on:click-outside={toggleFloatingElement}
        on:escape={toggleFloatingElement}
        slot="floating-element"
      >
        <MenuItem
          on:select={() => {
            toggleFloatingElement();
            onExport(FileExportType.Parquet);
          }}
        >
          Export as Parquet
        </MenuItem>
        <MenuItem
          on:select={() => {
            toggleFloatingElement();
            onExport(FileExportType.CSV);
          }}
        >
          Export as CSV
        </MenuItem>
      </Menu>
    </WithTogglableFloatingElement>
    <TooltipContent slot="tooltip-content">
      {#if modelHasError}Fix the errors in your model to export
      {:else}
        Export this model as a dataset
      {/if}
    </TooltipContent>
  </Tooltip>
  <CreateDashboardButton {width} hasError={modelHasError} {activeEntityID} />
</PanelCTA>

<div class="grow text-right px-4 pb-4 pt-2" style:height="56px">
  <!-- top row: row analysis -->
  <div
    class="flex flex-row items-center justify-between"
    class:text-gray-300={modelHasError}
  >
    <div
      class="italic text-gray-500"
      class:text-gray-500={modelHasError}
      class:italic={modelHasError}
    >
      <WithModelResultTooltip {modelHasError}>
        <div>
          {#if validRollup(rollup)}
            {#if isNaN(rollup)}
              ~
            {:else if rollup === 0}
              resultset is empty
            {:else if rollup !== 1}
              {formatBigNumberPercentage(rollup)}
              of source rows
            {:else}no change in row {#if containerWidth > COLUMN_PROFILE_CONFIG.hideRight}count{:else}ct.{/if}
            {/if}
          {:else if rollup === Infinity}
            &nbsp; {formatInteger(outputRowCardinalityValue)} row{#if outputRowCardinalityValue !== 1}s{/if}
            selected
          {/if}
        </div>

        <!-- tooltip content -->
        <svelte:fragment slot="tooltip-title">rollup percentage</svelte:fragment
        >
        <svelte:fragment slot="tooltip-description"
          >The ratio of resultset rows to source rows, as a percentage.</svelte:fragment
        >
      </WithModelResultTooltip>
    </div>
    <div
      class="text-gray-800 font-bold"
      class:font-normal={modelHasError}
      class:italic={modelHasError}
      class:text-gray-500={modelHasError}
    >
      {#if inputRowCardinalityValue > 0}
        {formatInteger(~~outputRowCardinalityValue)} row{#if outputRowCardinalityValue !== 1}s{/if}
      {:else if inputRowCardinalityValue === 0}
        no rows selected
      {:else}
        &nbsp;
      {/if}
    </div>
  </div>
  <!-- bottom row: column analysis -->

  <div class="flex flex-row justify-between">
    <WithModelResultTooltip {modelHasError}>
      <div
        class:font-normal={modelHasError}
        class:text-gray-500={modelHasError}
        class:italic={modelHasError}
      >
        {#if columnDelta > 0}
          {formatInteger(columnDelta)} column{#if columnDelta !== 1}s{/if} added
        {:else if columnDelta < 0}
          {formatInteger(-columnDelta)} column{#if -columnDelta !== 1}s{/if} dropped
        {:else if columnDelta === 0}
          no change in column count
        {:else}
          no change in column count
        {/if}
      </div>

      <!-- tooltip content -->
      <svelte:fragment slot="tooltip-title">column diff</svelte:fragment>
      <svelte:fragment slot="tooltip-description">
        The difference in column counts between the sources and model.</svelte:fragment
      >
    </WithModelResultTooltip>
    <div
      class="text-gray-800 font-bold"
      class:font-normal={modelHasError}
      class:text-gray-500={modelHasError}
      class:italic={modelHasError}
    >
      {currentDerivedModel?.profile?.length} columns
    </div>
  </div>
</div>
