<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceProfileColumns,
    useRuntimeServicePutFileAndReconcile,
    V1ReconcileResponse,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import CollapsibleSectionTitle from "@rilldata/web-local/lib/components/CollapsibleSectionTitle.svelte";
  import ColumnProfile from "@rilldata/web-local/lib/components/column-profile/ColumnProfile.svelte";
  import {
    GridCell,
    LeftRightGrid,
  } from "@rilldata/web-local/lib/components/left-right-grid";
  import PanelCTA from "@rilldata/web-local/lib/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-local/lib/components/panel/ResponsiveButtonText.svelte";
  import StickToHeaderDivider from "@rilldata/web-local/lib/components/panel/StickToHeaderDivider.svelte";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { selectTimestampColumnFromSchema } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getName } from "@rilldata/web-local/lib/util/incrementName";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { slide } from "svelte/transition";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { useCreateDashboardFromSource } from "../../../svelte-query/actions";
  import { useDashboardNames } from "../../../svelte-query/dashboards";
  import { useModelNames } from "../../../svelte-query/models";
  import { getSummaries } from "../../column-profile/queries";
  import { createModelFromSource } from "../../navigation/models/createModel";

  export let sourceName: string;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: getSource = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  let source: V1Source;
  $: source = $getSource?.data?.entry?.source;

  $: modelNames = useModelNames(runtimeInstanceId);
  $: dashboardNames = useDashboardNames(runtimeInstanceId);
  const createModelMutation = useRuntimeServicePutFileAndReconcile();
  const createDashboardFromSourceMutation = useCreateDashboardFromSource();

  let showColumns = true;

  // get source table references.

  // toggle state for inspector sections

  $: timestampColumns = selectTimestampColumnFromSchema(source?.schema);

  const handleCreateModelFromSource = async () => {
    const modelName = await createModelFromSource(
      queryClient,
      runtimeInstanceId,
      $modelNames.data,
      sourceName,
      $createModelMutation
    );
    navigationEvent.fireEvent(
      modelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Source,
      MetricsEventScreenName.Model
    );
  };

  const handleCreateDashboardFromSource = (sourceName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });
    const newModelName = getName(`${sourceName}_model`, $modelNames.data);
    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      $dashboardNames.data
    );
    $createDashboardFromSourceMutation.mutate(
      {
        data: {
          instanceId: $runtimeStore.instanceId,
          sourceName,
          newModelName,
          newDashboardName,
        },
      },
      {
        onSuccess: async (resp: V1ReconcileResponse) => {
          fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
          goto(`/dashboard/${newDashboardName}`);
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.RightPanel,
            MetricsEventScreenName.Source,
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(queryClient, runtimeInstanceId, resp);
        },
        onSettled: () => {
          overlay.set(null);
        },
      }
    );
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
      case "local_file":
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

  $: connectorType = formatConnectorType(source?.connector);
  $: fileExtension = getFileExtension(source);

  $: cardinalityQuery = useRuntimeServiceGetTableCardinality(
    $runtimeStore.instanceId,
    sourceName
  );
  $: cardinality = $cardinalityQuery?.data?.cardinality
    ? Number($cardinalityQuery?.data?.cardinality)
    : 0;

  /** get the current row count */
  $: {
    rowCount = `${formatInteger(cardinality)} row${
      cardinality !== 1 ? "s" : ""
    }`;
  }

  /** get the current column count */
  $: {
    columnCount = `${formatInteger(source?.schema?.fields?.length)} columns`;
  }

  /** total % null cells */

  $: profileColumns = useRuntimeServiceProfileColumns(
    $runtimeStore?.instanceId,
    sourceName,
    {},
    { query: { keepPreviousData: true } }
  );

  $: summaries = getSummaries(
    sourceName,
    $runtimeStore?.instanceId,
    $profileColumns?.data?.profileColumns
  );

  let totalNulls = undefined;

  $: if (summaries) {
    totalNulls = $summaries.reduce(
      (total, column) => total + (+column.nullCount || 0),
      0
    );
  }
  $: {
    const totalCells = source?.schema?.fields?.length * cardinality;
    nullPercentage = formatBigNumberPercentage(totalNulls / totalCells);
  }
</script>

<div class="table-profile">
  {#if source}
    <!-- CTAs -->
    <PanelCTA side="right" let:width>
      <Tooltip location="left" distance={16}>
        <Button type="secondary" on:click={handleCreateModelFromSource}>
          <ResponsiveButtonText {width}>Create Model</ResponsiveButtonText>
          <Model size="16px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          Create a model with these source columns
        </TooltipContent>
      </Tooltip>
      <Tooltip location="bottom" alignment="right" distance={16}>
        <Button
          type="primary"
          disabled={!timestampColumns?.length}
          on:click={() => handleCreateDashboardFromSource(sourceName)}
        >
          <ResponsiveButtonText {width}>Create Dashboard</ResponsiveButtonText>
          <Explore size="16px" />
        </Button>
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
          <GridCell side="left" classes="text-gray-600">
            {#if totalNulls !== undefined}
              {nullPercentage} null
            {/if}
          </GridCell>
          <TooltipContent slot="tooltip-content">
            {#if totalNulls !== undefined}
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

    <StickToHeaderDivider />

    <div class="pb-4 pt-4">
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="Source tables"
          bind:active={showColumns}
        >
          columns
        </CollapsibleSectionTitle>
      </div>

      {#if showColumns}
        <div transition:slide|local={{ duration: 200 }}>
          <ColumnProfile objectName={sourceName} indentLevel={0} />
        </div>
      {/if}
    </div>
  {/if}
</div>
