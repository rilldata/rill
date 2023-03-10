<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useEmbeddedSources } from "@rilldata/web-common/features/sources/selectors";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import {
    useQueryServiceTableCardinality,
    useQueryServiceTableColumns,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceListCatalogEntries,
    V1Model,
    V1TableCardinalityResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { derived } from "svelte/store";
  import { COLUMN_PROFILE_CONFIG } from "../../../../layout/config";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { getTableReferences } from "../../utils/get-table-references";
  import { getMatchingReferencesAndEntries } from "./utils";
  import WithModelResultTooltip from "./WithModelResultTooltip.svelte";

  export let modelName: string;
  export let containerWidth = 0;

  $: getModel = useRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );
  let model: V1Model;
  $: model = $getModel?.data?.entry?.model;

  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;

  let rollup: number;
  let sourceTableReferences;

  // get source table references.
  $: if (model?.sql) {
    sourceTableReferences = getTableReferences(model.sql);
  }

  $: embeddedSources = useEmbeddedSources($runtime.instanceId);

  // get the cardinalitie & table information.
  let cardinalityQueries: Array<UseQueryStoreResult<number>> = [];
  let sourceProfileColumns: Array<UseQueryStoreResult<number>> = [];

  $: getAllSources = useRuntimeServiceListCatalogEntries($runtime?.instanceId, {
    type: "OBJECT_TYPE_SOURCE",
  });

  $: getAllModels = useRuntimeServiceListCatalogEntries($runtime?.instanceId, {
    type: "OBJECT_TYPE_MODEL",
  });

  // for each reference, match to an existing model or source,
  $: referencedThings = getMatchingReferencesAndEntries(
    modelName,
    sourceTableReferences,
    [
      ...($getAllSources?.data?.entries || []),
      ...($getAllModels?.data?.entries || []),
    ]
  );
  $: if (sourceTableReferences?.length) {
    // first, pull out all references that are in the catalog.

    // then get the cardinalities.
    cardinalityQueries = referencedThings?.map(([entity, _reference]) => {
      return useQueryServiceTableCardinality(
        $runtime?.instanceId,
        entity.name,
        {},
        { query: { select: (data) => +data?.cardinality || 0 } }
      );
    });

    // then we'll get the total number of columns for comparison.
    sourceProfileColumns = referencedThings?.map(([entity]) => {
      return useQueryServiceTableColumns(
        $runtime?.instanceId,
        entity.name,
        {},
        { query: { select: (data) => data?.profileColumns?.length || 0 } }
      );
    });
  }

  // get input table cardinalities. We use this to determine the rollup factor.

  $: inputCardinalities = derived(cardinalityQueries, ($cardinalities) => {
    return $cardinalities
      .map((c: { data: number }) => c?.data)
      .reduce((total: number, cardinality: number) => total + cardinality, 0);
  });

  // get all source column amounts. We will use this determine the number of dropped columns.
  $: sourceColumns = derived(
    sourceProfileColumns,
    (columns) => {
      return columns
        .map((col) => col.data)
        .reduce((total: number, columns: number) => columns + total, 0);
    },
    0
  );

  let modelCardinalityQuery: UseQueryStoreResult<V1TableCardinalityResponse>;
  $: if (model?.name)
    modelCardinalityQuery = useQueryServiceTableCardinality(
      $runtime.instanceId,
      model?.name
    );
  let outputRowCardinalityValue: number;
  $: outputRowCardinalityValue = Number(
    $modelCardinalityQuery?.data?.cardinality ?? 0
  );

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
