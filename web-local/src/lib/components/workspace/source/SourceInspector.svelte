<script lang="ts">
  import {
    useRuntimeServiceGetCatalogObject,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import type { DerivedTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
  import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import { Button } from "@rilldata/web-local/lib/components/button";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import Explore from "@rilldata/web-local/lib/components/icons/Explore.svelte";
  import Model from "@rilldata/web-local/lib/components/icons/Model.svelte";
  import {
    GridCell,
    LeftRightGrid,
  } from "@rilldata/web-local/lib/components/left-right-grid";
  import PanelCTA from "@rilldata/web-local/lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import StickToHeaderDivider from "@rilldata/web-local/lib/components/panel/StickToHeaderDivider.svelte";
  import Tooltip from "@rilldata/web-local/lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-local/lib/components/tooltip/TooltipContent.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import {
    autoCreateMetricsDefinitionForSource,
    createModelForSource,
  } from "@rilldata/web-local/lib/redux-store/source/source-apis";
  import { selectTimestampColumnFromProfileEntity } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-local/lib/util/formatters";
  import { getContext } from "svelte";
  import { slide } from "svelte/transition";

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;
  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  export let sourceName: string;

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: getSource = useRuntimeServiceGetCatalogObject(
    runtimeInstanceId,
    sourceName
  );

  let showColumns = true;

  let currentTable: PersistentTableEntity;
  $: currentTable =
    sourceName && $persistentTableStore?.entities
      ? $persistentTableStore.entities.find((q) => q.tableName === sourceName)
      : undefined;
  let currentDerivedTable: DerivedTableEntity;
  $: currentDerivedTable =
    currentTable && $derivedTableStore?.entities
      ? $derivedTableStore.entities.find((q) => q.id === currentTable?.id)
      : undefined;
  // get source table references.

  // toggle state for inspector sections

  $: timestampColumns =
    selectTimestampColumnFromProfileEntity(currentDerivedTable);

  const handleCreateModelFromSource = async () => {
    const asynchronous = true;
    createModelForSource(
      $persistentModelStore.entities,
      currentTable.tableName,
      asynchronous
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
        (table) => table.tableName === sourceName
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
  let rowCount;
  let columnCount;
  let nullPercentage;

  function formatConnectorType(connectorType: string) {
    switch (connectorType) {
      case "s3":
        return "S3";
      case "gcs":
        return "GCS";
      case "https":
        return "http(s)";
      case "file":
        return "Local file";
      default:
        return "";
    }
  }

  function getFileExtension(source: V1Source): string {
    const path = source?.properties?.path?.toLowerCase();
    if (path?.includes(".csv")) return "CSV";
    if (path?.includes(".parquet")) return "Parquet";
    return "";
  }

  $: connectorType = formatConnectorType(
    $getSource.data?.object?.source?.connector
  );
  $: fileExtension = getFileExtension($getSource.data?.object?.source);

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
        <GridCell side="left"
          >{connectorType}
          {fileExtension !== "" ? `(${fileExtension})` : ""}</GridCell
        >
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
          <ColumnProfile
            objectName={sourceName}
            entityId={currentTable.id}
            indentLevel={0}
            cardinality={currentDerivedTable?.cardinality ?? 0}
            profile={currentDerivedTable?.profile ?? []}
            head={currentDerivedTable?.preview ?? []}
          />
        </div>
      {/if}
    </div>
  {/if}
</div>
