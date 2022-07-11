<script lang="ts">
  import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
  import { getContext, onMount, tick } from "svelte";
  import { sineOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { slide } from "svelte/transition";

  import Export from "$lib/components/icons/Export.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import * as classes from "$lib/util/component-classes";
  import { onClickOutside } from "$lib/util/on-click-outside";

  import ExportError from "$lib/components/modal/ExportError.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  import {
    ApplicationStore,
    config as appConfig,
    dataModelerService,
  } from "$lib/application-state-stores/application-store";

  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "$lib/util/formatters";

  import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
  import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "$lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "$lib/application-state-stores/table-stores";
  import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";

  import { FileExportType } from "$common/data-modeler-service/ModelActions";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import Button from "$lib/components/Button.svelte";
  import ColumnProfileNavEntry from "$lib/components/column-profile/ColumnProfileNavEntry.svelte";

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

  const store = getContext("rill:app:store") as ApplicationStore;
  const queryHighlight = getContext("rill:app:query-highlight");

  let rollup;
  let tables;
  // get source tables?
  let sourceTableReferences;
  let showColumns = true;

  let showExportErrorModal: boolean;
  let exportErrorMessage: string;

  // interface tweens for the  big numbers
  let bigRollupNumber = tweened(0, { duration: 700, easing });
  let outputRowCardinality = tweened(0, { duration: 250, easing });

  /** Select the explicit ID to prevent unneeded reactive updates in currentModel */
  $: activeEntityID = $store?.activeEntity?.id;

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
  $: if (
    outputRowCardinalityValue !== 0 &&
    outputRowCardinalityValue !== undefined
  ) {
    outputRowCardinality.set(outputRowCardinalityValue);
  }
  $: inputRowCardinalityValue = tables?.length
    ? tables.reduce((acc, v) => acc + v.cardinality, 0)
    : 0;
  $: if (
    inputRowCardinalityValue !== undefined &&
    outputRowCardinalityValue !== undefined
  ) {
    rollup = outputRowCardinalityValue / inputRowCardinalityValue;
  }

  function validRollup(number) {
    return rollup !== Infinity && rollup !== -Infinity && !isNaN(number);
  }

  $: if (rollup !== undefined && !isNaN(rollup)) bigRollupNumber.set(rollup);

  // toggle state for inspector sections
  let showSourceTables = true;

  let container;
  let containerWidth = 0;
  let contextMenu;
  let contextMenuOpen = false;
  let menuX;
  let menuY;
  let clickOutsideListener;
  $: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

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
      exportErrorMessage = `Failed to export.\n${exportResp.messages
        .map((message) => message.message)
        .join("\n")}`;
      showExportErrorModal = true;
    }
  };

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container.clientWidth;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });
</script>

{#key currentModel?.id}
  <div bind:this={container}>
    {#if currentModel && currentModel.query.trim().length && tables}
      <div
        style:height="var(--header-height)"
        class:text-gray-300={currentDerivedModel?.error}
        class="cost pl-4 pr-4 flex flex-row items-center gap-x-2"
      >
        {#if !currentDerivedModel?.error && rollup !== undefined && rollup !== Infinity && rollup !== -Infinity}
          <Tooltip
            location="left"
            alignment="middle"
            distance={16}
            suppress={contextMenuOpen}
          >
            <Button
              onClick={async (event) => {
                contextMenuOpen = !contextMenuOpen;
                menuX = event.clientX;
                menuY = event.clientY;
                if (!clickOutsideListener) {
                  await tick();
                  clickOutsideListener = onClickOutside(() => {
                    contextMenuOpen = false;
                  }, contextMenu);
                }
              }}
            >
              export
              <Export size="16px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              export this model as a dataset
            </TooltipContent>
          </Tooltip>

          <div class="grow text-right">
            <div
              class="cost-estimate text-gray-900 font-bold"
              class:text-gray-300={currentDerivedModel?.error}
            >
              {#if inputRowCardinalityValue > 0}
                {formatInteger(~~outputRowCardinalityValue)} row{#if outputRowCardinalityValue !== 1}s{/if}{#if containerWidth > COLUMN_PROFILE_CONFIG.hideRight},
                  {currentDerivedModel?.profile?.length} columns
                {/if}
              {:else if inputRowCardinalityValue === 0}
                no rows selected
              {:else}
                &nbsp;
              {/if}
            </div>
            <Tooltip location="left" alignment="center" distance={8}>
              <div class=" text-gray-500">
                {#if validRollup(rollup)}
                  {#if isNaN(rollup)}
                    ~
                  {:else if rollup === 0}
                    <!-- show no additional text. -->
                    resultset is empty
                  {:else if rollup !== 1}
                    {formatBigNumberPercentage(
                      rollup < 0.0005 ? rollup : $bigRollupNumber || 0
                    )}
                    of source rows
                  {:else}no change in row {#if containerWidth > COLUMN_PROFILE_CONFIG.hideRight}count{:else}ct.{/if}
                  {/if}
                {:else if rollup === Infinity}
                  &nbsp; {formatInteger(outputRowCardinalityValue)} row{#if outputRowCardinalityValue !== 1}s{/if}
                  selected
                {/if}
              </div>
              <TooltipContent slot="tooltip-content">
                <div class="pt-1 pb-1 font-bold">the rollup percentage</div>
                <div style:width="240px" class="pb-1">
                  the ratio of resultset rows to source rows, as a percentage
                </div>
              </TooltipContent>
            </Tooltip>
          </div>
        {/if}
      </div>
    {/if}

    <hr />

    <div class="model-profile">
      {#if currentModel && currentModel.query.trim().length}
        <div class="pt-4 pb-4">
          <div class=" pl-4 pr-4">
            <CollapsibleSectionTitle
              tooltipText="sources"
              bind:active={showSourceTables}
            >
              Sources
            </CollapsibleSectionTitle>
          </div>
          {#if showSourceTables}
            <div transition:slide|local={{ duration: 200 }} class="mt-1">
              {#if sourceTableReferences?.length && tables}
                {#each sourceTableReferences as reference, index (reference.name)}
                  {@const correspondingTableCardinality =
                    tables[index]?.cardinality}
                  <div
                    class="grid justify-between gap-x-2 {classes.QUERY_REFERENCE_TRIGGER} p-1 pl-4 pr-4"
                    style:grid-template-columns="auto max-content"
                    on:focus={() => {
                      queryHighlight.set(reference.tables);
                    }}
                    on:mouseover={() => {
                      queryHighlight.set(reference.tables);
                    }}
                    on:mouseleave={() => {
                      queryHighlight.set(undefined);
                    }}
                    on:blur={() => {
                      queryHighlight.set(undefined);
                    }}
                  >
                    <div
                      class="text-ellipsis overflow-hidden whitespace-nowrap"
                    >
                      {reference.name}
                    </div>
                    <div class="text-gray-500 italic">
                      <!-- is there a source table with this name and cardinality established? -->
                      {#if correspondingTableCardinality}
                        {`${formatInteger(
                          correspondingTableCardinality
                        )} rows` || ""}
                      {/if}
                    </div>
                  </div>
                {/each}
              {:else}
                <div class="pl-4 pr-5 p-1 italic text-gray-400">
                  none selected
                </div>
              {/if}
            </div>
          {/if}
        </div>

        <hr />

        <div class="pb-4 pt-4">
          <div class=" pl-4 pr-4">
            <CollapsibleSectionTitle
              tooltipText="source tables"
              bind:active={showColumns}
            >
              selected columns
            </CollapsibleSectionTitle>
          </div>

          {#if currentDerivedModel?.profile && showColumns}
            <div transition:slide|local={{ duration: 200 }}>
              <CollapsibleTableSummary
                entityType={EntityType.Model}
                showTitle={false}
                showContextButton={false}
                show={showColumns}
                name={currentModel.name}
                cardinality={currentDerivedModel?.cardinality ?? 0}
                active={currentModel?.id === $store?.activeEntity?.id}
              >
                <svelte:fragment slot="summary" let:containerWidth>
                  <ColumnProfileNavEntry
                    indentLevel={1}
                    {containerWidth}
                    cardinality={currentDerivedModel?.cardinality ?? 0}
                    profile={currentDerivedModel?.profile ?? []}
                    head={currentDerivedModel?.preview ?? []}
                  />
                </svelte:fragment>
              </CollapsibleTableSummary>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
{/key}

{#if contextMenuOpen}
  <!-- place this above codemirror.-->
  <div bind:this={contextMenu}>
    <FloatingElement
      relationship="mouse"
      target={{ x: menuX, y: menuY }}
      location="left"
      alignment="start"
    >
      <Menu
        color="dark"
        on:escape={() => {
          contextMenuOpen = false;
        }}
        on:item-select={() => {
          contextMenuOpen = false;
        }}
      >
        <MenuItem on:select={() => onExport(FileExportType.Parquet)}>
          Export as Parquet
        </MenuItem>
        <MenuItem on:select={() => onExport(FileExportType.CSV)}>
          Export as CSV
        </MenuItem>
      </Menu>
    </FloatingElement>
  </div>
  <ExportError bind:exportErrorMessage bind:showExportErrorModal />
{/if}

<style lang="postcss">
  .results {
    overflow: auto;
    max-width: var(--right-sidebar-width);
  }
</style>
