<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import {
    ColumnSummary,
    getSummaries,
  } from "@rilldata/web-common/components/column-profile/queries";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    formatConnectorType,
    getFileExtension,
  } from "@rilldata/web-common/features/sources/inspector/helpers";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import {
    createQueryServiceTableCardinality,
    createQueryServiceTableColumns,
    V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import { slide } from "svelte/transition";
  import { GridCell, LeftRightGrid } from "../../../components/grid";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { resourceIsLoading } from "../../entity-management/resource-selectors";
  import { useIsSourceUnsaved, useSource } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  $: runtimeInstanceId = $runtime.instanceId;

  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2 | undefined;
  $: source = $sourceQuery.data?.source;

  let showColumns = true;

  // get source table references.

  // toggle state for inspector sections

  /** source summary information */
  let rowCount;
  let columnCount;
  let nullPercentage: number | undefined;

  $: connectorType = source && formatConnectorType(source);
  $: fileExtension = source && getFileExtension(source);

  $: cardinalityQuery = createQueryServiceTableCardinality(
    $runtime.instanceId,
    sourceName,
  );
  $: cardinality = $cardinalityQuery?.data?.cardinality
    ? Number($cardinalityQuery?.data?.cardinality)
    : 0;

  $: profileColumns = createQueryServiceTableColumns(
    $runtime?.instanceId,
    sourceName,
    {},
    { query: { keepPreviousData: true } },
  );
  $: profileColumnsCount = $profileColumns?.data?.profileColumns?.length ?? 0;

  /** get the current row count */
  $: {
    rowCount = `${formatInteger(cardinality)} row${
      cardinality !== 1 ? "s" : ""
    }`;
  }

  /** get the current column count */
  $: {
    columnCount = `${formatInteger(profileColumnsCount)} columns`;
  }

  /** total % null cells */

  let summaries: Readable<Array<ColumnSummary>>;
  $: if ($profileColumns?.data?.profileColumns) {
    summaries = getSummaries(sourceName, $runtime?.instanceId, $profileColumns);
  }

  let totalNulls: number | undefined = undefined;

  $: if (summaries) {
    totalNulls = $summaries.reduce(
      (total, column) => total + (column?.nullCount ?? 0),
      0,
    );
  }
  $: {
    const totalCells = profileColumnsCount * cardinality;
    nullPercentage =
      totalNulls !== undefined
        ? formatBigNumberPercentage(totalNulls / totalCells)
        : undefined;
  }

  const sourceStore = useSourceStore(sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML,
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

<div class="{isSourceUnsaved && 'grayscale'} transition duration-200">
  {#if resourceIsLoading($sourceQuery?.data)}
    <div class="mt-6">
      <ReconcilingSpinner />
    </div>
  {:else if source && !$sourceQuery.isError}
    <!-- summary info -->
    <div class="p-4 pt-2">
      <LeftRightGrid>
        <GridCell side="left"
          >{connectorType}
          {fileExtension !== "" ? `(${fileExtension})` : ""}</GridCell
        >
        <GridCell side="right" classes="text-gray-800 font-bold">
          {rowCount}
        </GridCell>

        <Tooltip location="left" alignment="start" distance={24}>
          <GridCell side="left" classes="text-gray-600">
            {#if nullPercentage !== undefined}
              {nullPercentage} null
            {/if}
          </GridCell>
          <TooltipContent slot="tooltip-content">
            {#if nullPercentage !== undefined}
              {nullPercentage} of table values are null
            {:else}
              awaiting calculation of total null table values
            {/if}
          </TooltipContent>
        </Tooltip>
        <GridCell side="right" classes="text-gray-800 font-bold">
          {columnCount}
        </GridCell>
      </LeftRightGrid>
    </div>

    <hr />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="available columns"
          bind:active={showColumns}
        >
          columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns && source?.state?.table}
        <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
          <ColumnProfile objectName={source?.state?.table} indentLevel={0} />
        </div>
      {/if}
    </div>
  {/if}
</div>
