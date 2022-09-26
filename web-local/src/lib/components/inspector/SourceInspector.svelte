<script lang="ts">
  import type { DerivedTableEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { PersistentTableEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import { BehaviourEventMedium } from "$web-local/common/metrics-service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "$web-local/common/metrics-service/MetricsTypes";
  import type { ApplicationStore } from "../../application-state-stores/application-store";
  import type { PersistentModelStore } from "../../application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "../../application-state-stores/table-stores";
  import { Button } from "../button";
  import CollapsibleSectionTitle from "../CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "../column-profile/CollapsibleTableSummary.svelte";
  import ColumnProfileNavEntry from "../column-profile/ColumnProfileNavEntry.svelte";
  import Explore from "../icons/Explore.svelte";
  import Model from "../icons/Model.svelte";
  import { GridCell, LeftRightGrid } from "../left-right-grid";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import {
    autoCreateMetricsDefinitionForSource,
    createModelForSource,
  } from "../../redux-store/source/source-apis";
  import { selectTimestampColumnFromProfileEntity } from "../../redux-store/source/source-selectors";
  import { TableSourceType } from "../../types";
  import { navigationEvent } from "../../metrics/initMetrics";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "../../util/formatters";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";

  import PanelCTA from "../panel/PanelCTA.svelte";
  import ResponsiveButtonText from "../panel/ResponsiveButtonText.svelte";
  import StickToHeaderDivider from "../panel/StickToHeaderDivider.svelte";

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  const store = getContext("rill:app:store") as ApplicationStore;

  let showColumns = true;

  /** Select the explicit ID to prevent unneeded reactive updates in currentTable */
  $: activeEntityID = $store?.activeEntity?.id;

  let currentTable: PersistentTableEntity;
  $: currentTable =
    activeEntityID && $persistentTableStore?.entities
      ? $persistentTableStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  let currentDerivedTable: DerivedTableEntity;
  $: currentDerivedTable =
    activeEntityID && $derivedTableStore?.entities
      ? $derivedTableStore.entities.find((q) => q.id === activeEntityID)
      : undefined;
  // get source table references.

  // toggle state for inspector sections

  $: timestampColumns =
    selectTimestampColumnFromProfileEntity(currentDerivedTable);

  const handleCreateModelFromSource = async () => {
    createModelForSource(
      $persistentModelStore.entities,
      currentTable.tableName
    ).then((createdModelId) => {
      navigationEvent.fireEvent(
        createdModelId,
        BehaviourEventMedium.Button,
        MetricsEventSpace.RightPanel,
        MetricsEventScreenName.Source,
        MetricsEventScreenName.Model
      );
    });
  };

  const handleCreateMetric = () => {
    // A side effect of the createMetricsDefsApi is we switch active assets to
    // the newly created metrics definition. So, this'll bring us to the
    // MetricsDefinition page. (The logic for this is contained in the
    // not-pictured async thunk.)
    autoCreateMetricsDefinitionForSource(
      $persistentModelStore.entities,
      $derivedTableStore.entities,
      currentTable.id,
      $persistentTableStore.entities.find(
        (table) => table.id === activeEntityID
      ).tableName
    ).then((createdMetricsId) => {
      navigationEvent.fireEvent(
        createdMetricsId,
        BehaviourEventMedium.Button,
        MetricsEventSpace.RightPanel,
        MetricsEventScreenName.Source,
        MetricsEventScreenName.Dashboard
      );
    });
  };

  /** source summary information */
  let sourceType;
  let rowCount;
  let columnCount;
  let nullPercentage;
  $: {
    switch (currentTable?.sourceType) {
      case TableSourceType.ParquetFile: {
        sourceType = "Parquet";
        break;
      }
      case TableSourceType.CSVFile: {
        sourceType = `CSV (${currentTable?.csvDelimiter || "comma"})`;
        break;
      }
      case TableSourceType.DuckDB: {
        sourceType = "DuckDB";
        break;
      }
      default: {
        sourceType = "unknown";
        break;
      }
    }
  }

  /** get the current row count */
  $: {
    rowCount = `${formatInteger(currentDerivedTable?.cardinality)} row${
      currentDerivedTable?.cardinality !== 1 ? "s" : ""
    }`;
  }

  /** get the current column count */
  $: {
    columnCount = `${formatInteger(
      currentDerivedTable?.profile?.length
    )} columns`;
  }

  /** total % null cells */

  $: {
    const totalCells =
      currentDerivedTable?.profile?.length * currentDerivedTable?.cardinality;
    const totalNulls = currentDerivedTable?.profile
      .map((profile) => profile?.nullCount)
      .reduce((total, count) => total + count, 0);
    nullPercentage = formatBigNumberPercentage(totalNulls / totalCells);
  }
</script>

<div class="table-profile">
  {#if currentTable}
    <!-- CTAs -->
    <PanelCTA side="right" let:width>
      <Tooltip location="left" distance={16}>
        <Button type="secondary" on:click={handleCreateModelFromSource}>
          <ResponsiveButtonText {width}>Create Model</ResponsiveButtonText>
          <Model size="16px" /></Button
        >
        <TooltipContent slot="tooltip-content">
          Create a model with these source columns
        </TooltipContent>
      </Tooltip>
      <Tooltip location="bottom" alignment="right" distance={16}>
        <Button
          type="primary"
          disabled={!timestampColumns?.length}
          on:click={handleCreateMetric}
        >
          <ResponsiveButtonText {width}>Create Dashboard</ResponsiveButtonText>
          <Explore size="16px" /></Button
        >
        <TooltipContent slot="tooltip-content">
          {#if timestampColumns?.length}
            Auto create metrics based on your data source and go to dashboard
          {:else}
            This data source does not have a TIMESTAMP column
          {/if}
        </TooltipContent>
      </Tooltip>
    </PanelCTA>

    <!-- summary info -->
    <div class=" p-4 pt-2">
      <LeftRightGrid>
        <GridCell side="left">
          {sourceType}
        </GridCell>
        <GridCell side="right" classes="text-gray-800 font-bold">
          {rowCount}
        </GridCell>

        <Tooltip location="left" alignment="start" distance={24}>
          <GridCell side="left" classes="text-gray-600 italic">
            {nullPercentage} null
          </GridCell>
          <TooltipContent slot="tooltip-content">
            {nullPercentage} of table values are null
          </TooltipContent>
        </Tooltip>
        <GridCell side="right" classes="text-gray-800 font-bold">
          {columnCount}
        </GridCell>
      </LeftRightGrid>
    </div>

    <StickToHeaderDivider />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="source tables"
          bind:active={showColumns}
        >
          columns
        </CollapsibleSectionTitle>
      </div>

      {#if currentDerivedTable?.profile && showColumns}
        <div transition:slide|local={{ duration: 200 }}>
          <CollapsibleTableSummary
            entityType={EntityType.Table}
            showTitle={false}
            show={showColumns}
            name={currentTable.name}
            cardinality={currentDerivedTable?.cardinality ?? 0}
            active={currentTable?.id === activeEntityID}
          >
            <svelte:fragment slot="summary" let:containerWidth>
              <ColumnProfileNavEntry
                entityId={currentTable.id}
                indentLevel={0}
                {containerWidth}
                cardinality={currentDerivedTable?.cardinality ?? 0}
                profile={currentDerivedTable?.profile ?? []}
                head={currentDerivedTable?.preview ?? []}
              />
            </svelte:fragment>
          </CollapsibleTableSummary>
        </div>
      {/if}
    </div>
  {/if}
</div>
