<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    useRuntimeServiceListCatalogObjects,
    useRuntimeServiceMigrateDelete,
    useRuntimeServiceMigrateSingle,
    useRuntimeServiceTriggerRefresh,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
  import { refreshSource } from "@rilldata/web-local/lib/components/navigation/sources/refreshSource";
  import { getContext } from "svelte";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import {
    ApplicationStore,
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import type { PersistentModelStore } from "../../../application-state-stores/model-stores";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "../../../application-state-stores/table-stores";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import {
    autoCreateMetricsDefinitionForSource,
    createModelForSource,
    sourceUpdated,
  } from "../../../redux-store/source/source-apis";
  import { derivedProfileEntityHasTimestampColumn } from "../../../redux-store/source/source-selectors";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import CollapsibleSectionTitle from "../../CollapsibleSectionTitle.svelte";
  import CollapsibleTableSummary from "../../column-profile/CollapsibleTableSummary.svelte";
  import ColumnProfileNavEntry from "../../column-profile/ColumnProfileNavEntry.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";
  import Add from "../../icons/Add.svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import Import from "../../icons/Import.svelte";
  import Model from "../../icons/Model.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import Source from "../../icons/Source.svelte";
  import { Divider, MenuItem } from "../../menu";
  import RenameAssetModal from "../RenameAssetModal.svelte";
  import AddSourceModal from "./AddSourceModal.svelte";

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  const persistentTableStore = getContext(
    "rill:app:persistent-table-store"
  ) as PersistentTableStore;

  const derivedTableStore = getContext(
    "rill:app:derived-table-store"
  ) as DerivedTableStore;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  const applicationStore = getContext("rill:app:store") as ApplicationStore;

  let showTables = true;

  let showAddSourceModal = false;

  const openShowAddSourceModal = () => {
    showAddSourceModal = true;
  };

  let showRenameTableModal = false;
  let renameTableID = null;
  let renameTableName = null;

  const openRenameTableModal = (tableID: string, tableName: string) => {
    showRenameTableModal = true;
    renameTableID = tableID;
    renameTableName = tableName;
  };

  const queryHandler = async (tableName: string) => {
    const asynchronous = true;
    await createModelForSource(
      $persistentModelStore.entities,
      tableName,
      asynchronous
    );
  };

  const quickStartMetrics = async (id: string, tableName: string) => {
    const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    const createdMetricsId = await autoCreateMetricsDefinitionForSource(
      $persistentModelStore.entities,
      $derivedTableStore.entities,
      id,
      tableName
    );

    navigationEvent.fireEvent(
      createdMetricsId,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      EntityTypeToScreenMap[previousActiveEntity],
      MetricsEventScreenName.Dashboard
    );
  };

  const viewSource = (id: string) => {
    goto(`/source/${id}`);

    if (id != activeEntityID) {
      const previousActiveEntity = $rillAppStore?.activeEntity?.type;
      navigationEvent.fireEvent(
        id,
        BehaviourEventMedium.AssetName,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Source
      );
    }
  };

  const deleteSource = useRuntimeServiceMigrateDelete();

  const handleDeleteSource = (tableName: string, id: string) => {
    $deleteSource.mutate(
      {
        instanceId: runtimeInstanceId,
        data: {
          name: tableName.toLowerCase(),
        },
      },
      {
        onSuccess: () => {
          if (
            $applicationStore.activeEntity.type === EntityType.Table &&
            $applicationStore.activeEntity.id === id
          ) {
            const nextSourceId = getNextEntityId(
              $persistentTableStore.entities,
              id
            );
            if (nextSourceId) {
              goto(`/source/${nextSourceId}`);
            } else {
              goto("/");
            }
          }
          sourceUpdated(tableName);
        },
      }
    );
  };

  const createModel = (tableName: string) => {
    const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    const asynchronous = true;

    createModelForSource(
      $persistentModelStore.entities,
      tableName,
      asynchronous
    ).then((createdModelId) => {
      navigationEvent.fireEvent(
        createdModelId,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Model
      );
    });
  };

  $: runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createSource = useRuntimeServiceMigrateSingle();
  $: getSources = useRuntimeServiceListCatalogObjects(runtimeInstanceId);

  const onRefreshSource = async (id: string, tableName: string) => {
    try {
      await refreshSource(
        $getSources.data?.objects.find(
          (object) => object.source?.name === tableName
        )?.source.connector,
        tableName,
        $runtimeStore,
        $refreshSourceMutation,
        $createSource
      );
      // invalidate the data preview (async)
      dataModelerService.dispatch("collectTableInfo", [id]);

      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogObjectQueryKey(
        runtimeInstanceId,
        tableName
      );
      await queryClient.invalidateQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

  $: activeEntityID = $rillAppStore?.activeEntity?.id;
</script>

<div
  class="pl-4 pb-3 pr-4 pt-5 grid justify-between"
  style="grid-template-columns: auto max-content;"
>
  <CollapsibleSectionTitle tooltipText={"sources"} bind:active={showTables}>
    <h4 class="flex flex-row items-center gap-x-2">
      <Source size="16px" /> Sources
    </h4>
  </CollapsibleSectionTitle>
  <ContextButton
    id={"create-table-button"}
    tooltipText="import csv or parquet file as a source"
    on:click={openShowAddSourceModal}
  >
    <Add />
  </ContextButton>
</div>
{#if showTables}
  <div class="pb-6" transition:slide|local={{ duration: 200 }}>
    {#if $persistentTableStore?.entities && $derivedTableStore?.entities}
      <!-- TODO: fix the object property access back to t.id from t["id"] once svelte fixes it -->
      {#each $persistentTableStore.entities as { tableName, id } (id)}
        {@const derivedTable = $derivedTableStore.entities.find(
          (t) => t["id"] === id
        )}
        {@const entityIsActive = id === activeEntityID}
        {@const source = $getSources.data?.objects?.find(
          (object) => object.source?.name === tableName
        )?.source}
        <div animate:flip={{ duration: 200 }} out:slide={{ duration: 200 }}>
          <CollapsibleTableSummary
            on:query={() => queryHandler(tableName)}
            on:select={() => viewSource(id)}
            entityType={EntityType.Table}
            name={tableName}
            cardinality={derivedTable?.cardinality ?? 0}
            sizeInBytes={derivedTable?.sizeInBytes ?? 0}
            active={entityIsActive}
            loading={$refreshSourceMutation.isLoading}
          >
            <ColumnProfileNavEntry
              slot="summary"
              let:containerWidth
              indentLevel={1}
              {containerWidth}
              cardinality={derivedTable?.cardinality ?? 0}
              profile={derivedTable?.profile ?? []}
              head={derivedTable?.preview ?? []}
              entityId={id}
            />
            <svelte:fragment slot="menu-items" let:toggleMenu>
              <MenuItem icon on:select={() => createModel(tableName)}>
                <Model slot="icon" />
                create new model
              </MenuItem>

              <MenuItem
                disabled={!derivedProfileEntityHasTimestampColumn(derivedTable)}
                icon
                on:select={() => quickStartMetrics(id, tableName)}
              >
                <Explore slot="icon" />
                autogenerate dashboard
                <svelte:fragment slot="description">
                  {#if !derivedProfileEntityHasTimestampColumn(derivedTable)}
                    requires a timestamp column
                  {/if}
                </svelte:fragment>
              </MenuItem>

              {#if source.connector === "file"}
                <MenuItem icon on:select={() => onRefreshSource(id, tableName)}>
                  <svelte:fragment slot="icon">
                    <Import />
                  </svelte:fragment>
                  import local file to refresh source
                </MenuItem>
              {:else}
                <MenuItem icon on:select={() => onRefreshSource(id, tableName)}>
                  <svelte:fragment slot="icon">
                    <RefreshIcon />
                  </svelte:fragment>
                  refresh source data
                </MenuItem>
              {/if}

              <Divider />
              <MenuItem
                icon
                on:select={() => {
                  openRenameTableModal(id, tableName);
                }}
              >
                <EditIcon slot="icon" />

                rename...
              </MenuItem>
              <!-- FIXME: this should pop up an "are you sure?" modal -->
              <MenuItem
                icon
                on:select={() => handleDeleteSource(tableName, id)}
              >
                <Cancel slot="icon" />
                delete</MenuItem
              >
            </svelte:fragment>
          </CollapsibleTableSummary>
        </div>
      {/each}
    {/if}
  </div>
  {#if showAddSourceModal}
    <AddSourceModal
      on:close={() => {
        showAddSourceModal = false;
      }}
    />
  {/if}
  {#if showRenameTableModal}
    <RenameAssetModal
      entityType={EntityType.Table}
      closeModal={() => (showRenameTableModal = false)}
      entityId={renameTableID}
      currentAssetName={renameTableName}
    />
  {/if}
{/if}
