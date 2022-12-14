<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceProfileColumns,
    V1GetTableCardinalityResponse,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import WithTogglableFloatingElement from "@rilldata/web-local/lib/components/floating-element/WithTogglableFloatingElement.svelte";
  import Export from "@rilldata/web-local/lib/components/icons/Export.svelte";
  import { Menu, MenuItem } from "@rilldata/web-local/lib/components/menu";
  import PanelCTA from "@rilldata/web-local/lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-local/lib/util/formatters";
  import { getTableReferences } from "@rilldata/web-local/lib/util/get-table-references";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { derived } from "svelte/store";
  import WithModelResultTooltip from "../WithModelResultTooltip.svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let modelName: string;
  export let containerWidth = 0;

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtimeStore.instanceId,
    modelName,
    { query: { queryKey: `current-model-query-in-inspector-${modelName}` } }
  );
  let model: V1Model;
  $: model = $getModel?.data?.entry?.model;

  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;

  let contextMenuOpen = false;

  const onExport = async (exportExtension: "csv" | "parquet") => {
    // TODO: how do we handle errors ?
    window.open(
      `${RuntimeUrl}/v1/instances/${$runtimeStore.instanceId}/table/${modelName}/export/${exportExtension}`
    );
  };

  let rollup;
  let sourceTableReferences;

  // get source table references.
  $: if (model?.sql) {
    sourceTableReferences = getTableReferences(model.sql);
  }

  // get the cardinalitie & table information.
  let cardinalityQueries = [];
  let sourceProfileColumns = [];
  $: if (sourceTableReferences?.length) {
    cardinalityQueries = sourceTableReferences.map((table) => {
      return useRuntimeServiceGetTableCardinality(
        $runtimeStore?.instanceId,
        table.reference,
        {},
        { query: { select: (data) => +data?.cardinality || 0 } }
      );
    });
    sourceProfileColumns = sourceTableReferences.map((table) => {
      return useRuntimeServiceProfileColumns(
        $runtimeStore?.instanceId,
        table.reference,
        {},
        { query: { select: (data) => data?.profileColumns?.length || 0 } }
      );
    });
  }

  // get input table cardinalities. We use this to determine the rollup factor.
  $: inputCardinalities = derived(cardinalityQueries, ($cardinalities) => {
    return $cardinalities
      .map((c: { data: number }) => c.data)
      .reduce((total: number, cardinality: number) => total + cardinality, 0);
  });

  // get all source column amounts. We will use this determine the number of dropped columns.
  $: sourceColumns = derived(
    sourceProfileColumns,
    ($columns) => {
      return $columns
        .map((col) => col.data)
        .reduce((total: number, columns: number) => columns + total, 0);
    },
    0
  );

  let modelCardinalityQuery: UseQueryStoreResult<V1GetTableCardinalityResponse>;
  $: if (model?.name)
    modelCardinalityQuery = useRuntimeServiceGetTableCardinality(
      $runtimeStore.instanceId,
      model?.name
    );
  $: outputRowCardinalityValue = $modelCardinalityQuery?.data?.cardinality;

  $: if (
    ($inputCardinalities !== undefined &&
      outputRowCardinalityValue !== undefined) ||
    $inputCardinalities
  ) {
    rollup = outputRowCardinalityValue / $inputCardinalities;
  }

  function validRollup(number) {
    return rollup !== Infinity && rollup !== -Infinity && !isNaN(number);
  }

  $: outputColumnNum = model?.schema?.fields?.length ?? 0;
  $: columnDelta = outputColumnNum - $sourceColumns;

  $: modelHasError = !!modelError;
</script>

<PanelCTA let:width side="right">
  <Tooltip
    alignment="middle"
    distance={16}
    location="left"
    suppress={contextMenuOpen}
  >
    <!-- attach floating element right here-->
    <WithTogglableFloatingElement
      alignment="start"
      bind:active={contextMenuOpen}
      distance={16}
      let:toggleFloatingElement
      location="left"
    >
      <Button
        disabled={modelHasError}
        on:click={toggleFloatingElement}
        type="secondary"
      >
        <ResponsiveButtonText {width}>Export Results</ResponsiveButtonText>
        <Export size="14px" />
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
            onExport("parquet");
          }}
        >
          Export as Parquet
        </MenuItem>
        <MenuItem
          on:select={() => {
            toggleFloatingElement();
            onExport("csv");
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
  <CreateDashboardButton hasError={modelHasError} {modelName} {width} />
</PanelCTA>

<div class="grow text-right px-4 pb-4 pt-2" style:height="56px">
  <!-- top row: row analysis -->
  <div
    class="flex flex-row items-center justify-between"
    class:text-gray-300={modelHasError}
  >
    <div class="text-gray-500" class:text-gray-500={modelHasError}>
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
      class="text-gray-800 ui-copy-strong"
      class:font-normal={modelHasError}
      class:text-gray-500={modelHasError}
    >
      {#if $inputCardinalities > 0}
        {formatInteger(~~outputRowCardinalityValue)} row{#if $inputCardinalities !== 1}s{/if}
      {:else if $inputCardinalities === 0}
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
      class="text-gray-800 ui-copy-strong"
      class:font-normal={modelHasError}
      class:text-gray-500={modelHasError}
    >
      {outputColumnNum} columns
    </div>
  </div>
</div>
