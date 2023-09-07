<script lang="ts">
  import { getColumnsProfileStore } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import {
    createRuntimeServiceGetCatalogEntry,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { COLUMN_PROFILE_CONFIG } from "../../../../layout/config";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  export let modelName: string;
  export let containerWidth = 0;

  $: getModel = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );
  let model: V1Model;
  $: model = $getModel?.data?.entry?.model;

  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;

  let rollup: number;

  const columnsProfile = getColumnsProfileStore();

  // get input table cardinalities. We use this to determine the rollup factor.
  $: inputCardinalities = $columnsProfile.references.reduce(
    (total, ref) => (ref?.cardinality ?? 0) + total,
    0
  );

  // get all source column amounts. We will use this determine the number of dropped columns.
  $: sourceColumns = $columnsProfile.references.reduce(
    (total, ref) => (ref?.columns?.length ?? 0) + total,
    0
  );

  let outputRowCardinalityValue: number;
  $: outputRowCardinalityValue = Number($columnsProfile.tableRows ?? 0);

  $: if (
    (inputCardinalities !== undefined &&
      outputRowCardinalityValue !== undefined) ||
    inputCardinalities
  ) {
    rollup = outputRowCardinalityValue / inputCardinalities;
  }

  function validRollup(number) {
    return rollup !== Infinity && rollup !== -Infinity && !isNaN(number);
  }

  $: outputColumnNum = model?.schema?.fields?.length ?? 0;
  $: columnDelta = outputColumnNum - sourceColumns;

  $: modelHasError = !!modelError;
</script>

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
              Result set is empty
            {:else if rollup !== 1}
              {formatBigNumberPercentage(rollup)}
              of source rows
            {:else}No change in row
              {#if containerWidth > COLUMN_PROFILE_CONFIG.hideRight}count{:else}ct.{/if}
            {/if}
          {:else if rollup === Infinity}
            {`${formatInteger(outputRowCardinalityValue)} row${
              outputRowCardinalityValue !== 1 ? "s" : ""
            } selected`}
          {/if}
        </div>

        <!-- tooltip content -->
        <svelte:fragment slot="tooltip-title"
          >Rollup percentage
        </svelte:fragment>
        <svelte:fragment slot="tooltip-description"
          >The ratio of resultset rows to source rows, as a percentage.
        </svelte:fragment>
      </WithModelResultTooltip>
    </div>
    <div
      class="text-gray-800 font-bold"
      class:font-normal={modelHasError}
      class:text-gray-500={modelHasError}
    >
      {#if outputRowCardinalityValue > 0}
        {`${formatInteger(outputRowCardinalityValue)} row${
          outputRowCardinalityValue !== 1 ? "s" : ""
        }`}
      {:else if outputRowCardinalityValue === 0}
        No rows selected
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
          {`${formatInteger(columnDelta)} column${
            columnDelta !== 1 ? "s" : ""
          } added`}
        {:else if columnDelta < 0}
          {`${formatInteger(-columnDelta)} column${
            -columnDelta !== 1 ? "s" : ""
          } dropped`}
        {:else if columnDelta === 0}
          No change in column count
        {:else}
          No change in column count
        {/if}
      </div>

      <!-- tooltip content -->
      <svelte:fragment slot="tooltip-title">Column diff</svelte:fragment>
      <svelte:fragment slot="tooltip-description">
        The difference in column counts between the sources and model.
      </svelte:fragment>
    </WithModelResultTooltip>
    <div
      class="text-gray-800 font-bold"
      class:font-normal={modelHasError}
      class:text-gray-500={modelHasError}
    >
      {outputColumnNum} columns
    </div>
  </div>
</div>
