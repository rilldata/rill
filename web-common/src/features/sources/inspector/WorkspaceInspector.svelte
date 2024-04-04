<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import { getSummaries } from "@rilldata/web-common/components/column-profile/queries";
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
    V1ModelV2,
    V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import { runtime } from "../../../runtime-client/runtime-store";
  import InspectorSummary from "./InspectorSummary.svelte";
  import { useModels } from "@rilldata/web-common/features/models/selectors";
  import { useSources } from "@rilldata/web-common/features/sources/selectors";
  import { derived } from "svelte/store";
  import { COLUMN_PROFILE_CONFIG } from "../../../layout/config";
  import { getTableReferences } from "../../models/utils/get-table-references";
  import { getMatchingReferencesAndEntries } from "../../models/workspace/inspector/utils";
  import WithModelResultTooltip from "../../models/workspace/inspector/WithModelResultTooltip.svelte";
  import References from "../../models/workspace/inspector/References.svelte";

  export let hasUnsavedChanges: boolean;
  export let tableName: string;
  export let source: V1SourceV2 | undefined = undefined;
  export let model: V1ModelV2 | undefined = undefined;
  export let sourceIsReconciling: boolean;
  export let containerWidth = 0;
  export let isEmpty = false;
  export let hasErrors: boolean;
  export let showReferences = true;
  export let showSummaryTitle = false;

  let showColumns = true;

  $: instanceId = $runtime.instanceId;

  $: connectorType = source && formatConnectorType(source);
  $: fileExtension = source && getFileExtension(source);

  $: cardinalityQuery = createQueryServiceTableCardinality(
    instanceId,
    tableName,
  );

  $: profileColumnsQuery = createQueryServiceTableColumns(
    instanceId,
    tableName,
    {},
    { query: { keepPreviousData: true } },
  );

  $: cardinality = Number($cardinalityQuery?.data?.cardinality ?? 0);

  $: profileColumnsCount =
    $profileColumnsQuery?.data?.profileColumns?.length ?? 0;

  $: rowCount = `${formatInteger(cardinality)} ${
    cardinality !== 1 ? "rows" : "row"
  }`;

  $: columnCount = `${formatInteger(profileColumnsCount)} columns`;

  $: summaries = getSummaries(tableName, instanceId, $profileColumnsQuery);

  $: totalCells = profileColumnsCount * cardinality;

  $: totalNulls = $summaries?.reduce(
    (total, column) => total + (column?.nullCount ?? 0),
    0,
  );

  $: nullPercentage =
    totalNulls !== undefined
      ? formatBigNumberPercentage(totalNulls / totalCells)
      : undefined;

  $: sourceTableReferences =
    model && getTableReferences(model?.spec?.sql ?? "");

  $: getAllSources = useSources(instanceId);
  $: getAllModels = useModels(instanceId);

  $: allSources = $getAllSources?.data ?? [];
  $: allModels = $getAllModels?.data ?? [];

  $: referencedThings =
    sourceTableReferences &&
    getMatchingReferencesAndEntries(tableName, sourceTableReferences, [
      ...allSources,
      ...allModels,
    ]);

  $: cardinalityQueries =
    referencedThings?.map(([resource]) => {
      return createQueryServiceTableCardinality(
        instanceId,
        resource.meta?.name?.name ?? "",
        {},
        { query: { select: (data) => +(data?.cardinality ?? 0) } },
      );
    }) ?? [];

  $: sourceProfileColumns =
    referencedThings?.map(([resource]) => {
      return createQueryServiceTableColumns(
        instanceId,
        resource.meta?.name?.name ?? "",
        {},
        { query: { select: (data) => data?.profileColumns?.length || 0 } },
      );
    }) ?? [];

  $: inputCardinalities = derived(cardinalityQueries, ($cardinalities) => {
    return $cardinalities
      .map((c) => c?.data ?? 0)
      .reduce((total: number, cardinality: number) => total + cardinality, 0);
  });

  $: sourceColumns = derived(
    sourceProfileColumns,
    (columns) =>
      columns
        .map((col) => col.data ?? 0)
        .reduce((total, columns) => columns + total, 0),

    0,
  );

  $: rollup = cardinality / $inputCardinalities;

  $: columnDelta = profileColumnsCount - $sourceColumns;
</script>

<div class="wrapper" class:grayscale={hasUnsavedChanges}>
  {#if sourceIsReconciling}
    <div class="size-full flex items-center justify-center">
      <ReconcilingSpinner />
    </div>
  {:else if isEmpty}
    <div class="px-4 py-24 italic ui-copy-disabled text-center">
      {source ? "Source" : "Model"} is empty.
    </div>
  {:else if source || model}
    <InspectorSummary {rowCount} {columnCount} showTitle={showSummaryTitle}>
      <svelte:fragment slot="row-header">
        {#if source}
          <p>
            {connectorType}
            {fileExtension && `(${fileExtension})`}
          </p>
        {:else}
          <WithModelResultTooltip modelHasError={hasErrors}>
            <p>
              {#if isNaN(rollup)}
                ~
              {:else if rollup === 0}
                Result set is empty
              {:else if rollup === Infinity}
                {rowCount} selected
              {:else if rollup !== 1}
                {formatBigNumberPercentage(rollup)}
                of source rows
              {:else}No change in row
                {containerWidth > COLUMN_PROFILE_CONFIG.hideRight
                  ? "count"
                  : "ct."}
              {/if}
            </p>

            <svelte:fragment slot="tooltip-title">
              Rollup percentage
            </svelte:fragment>
            <svelte:fragment slot="tooltip-description"
              >The ratio of resultset rows to source rows, as a percentage.
            </svelte:fragment>
          </WithModelResultTooltip>
        {/if}
      </svelte:fragment>

      <svelte:fragment slot="column-header">
        {#if source}
          <Tooltip location="left" alignment="start" distance={24}>
            {#if nullPercentage !== undefined}
              <p>{nullPercentage} null</p>
            {/if}

            <TooltipContent slot="tooltip-content">
              {#if nullPercentage !== undefined}
                {nullPercentage} of table values are null
              {:else}
                awaiting calculation of total null table values
              {/if}
            </TooltipContent>
          </Tooltip>
        {:else}
          <WithModelResultTooltip modelHasError={hasErrors}>
            <div class:font-normal={hasErrors} class:text-gray-500={hasErrors}>
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
        {/if}
      </svelte:fragment>
    </InspectorSummary>

    <hr />

    {#if showReferences && referencedThings?.length}
      <References modelHasError={hasErrors} {referencedThings} />
      <hr />
    {/if}

    <div>
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="available columns"
          bind:active={showColumns}
        >
          {model ? "Model columns" : "Source columns"}
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
          <ColumnProfile objectName={tableName} indentLevel={0} />
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply transition duration-200 py-2 flex flex-col gap-y-2;
  }
</style>
