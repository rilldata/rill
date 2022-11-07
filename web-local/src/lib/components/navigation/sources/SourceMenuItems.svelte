<script lang="ts">
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    useRuntimeServiceListCatalogObjects,
    useRuntimeServiceMigrateDelete,
    useRuntimeServiceMigrateSingle,
    useRuntimeServiceTriggerRefresh,
  } from "@rilldata/web-common/runtime-client";
  import type { ApplicationStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import { derivedProfileEntityHasTimestampColumn } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import { createEventDispatcher, getContext } from "svelte";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import Import from "../../icons/Import.svelte";
  import Model from "../../icons/Model.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";

  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";

  import { goto } from "$app/navigation";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getNextEntityId } from "@rilldata/web-local/common/utils/getNextEntityId";
  import {
    autoCreateMetricsDefinitionForSource,
    createModelForSource,
    sourceUpdated,
  } from "@rilldata/web-local/lib/redux-store/source/source-apis";
  import { Divider, MenuItem } from "../../menu";
  import { refreshSource } from "./refreshSource";

  export let sourceID;

  const dispatch = createEventDispatcher();

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

  const deleteSource = useRuntimeServiceMigrateDelete();
  $: runtimeInstanceId = $runtimeStore.instanceId;
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createSource = useRuntimeServiceMigrateSingle();
  $: getSources = useRuntimeServiceListCatalogObjects(runtimeInstanceId);

  $: persistentTable = $persistentTableStore?.entities?.find(
    (source) => source.id === sourceID
  );

  $: derivedTable = $derivedTableStore?.entities?.find(
    (source) => source.id === sourceID
  );
  $: currentSourceObject = $getSources?.data?.objects?.find(
    (object) => object.source.name === persistentTable.tableName
  );

  const handleDeleteSource = (tableName: string) => {
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
            $rillAppStore.activeEntity.type === EntityType.Table &&
            $rillAppStore.activeEntity.id === sourceID
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

  const bootstrapDashboard = async (id: string, tableName: string) => {
    const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    const createdMetricsId = await autoCreateMetricsDefinitionForSource(
      $persistentModelStore.entities,
      $derivedTableStore.entities,
      sourceID,
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

  const onRefreshSource = async (id: string, tableName: string) => {
    try {
      await refreshSource(
        currentSourceObject?.source?.connector,
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
</script>

<MenuItem icon on:select={() => createModel(currentSourceObject.source.name)}>
  <Model slot="icon" />
  create new model
</MenuItem>

<MenuItem
  disabled={!derivedProfileEntityHasTimestampColumn(derivedTable)}
  icon
  on:select={() =>
    bootstrapDashboard(sourceID, currentSourceObject.source.name)}
>
  <Explore slot="icon" />
  autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !derivedProfileEntityHasTimestampColumn(derivedTable)}
      requires a timestamp column
    {/if}
  </svelte:fragment>
</MenuItem>

{#if currentSourceObject?.source?.connector === "file"}
  <MenuItem
    icon
    on:select={() => onRefreshSource(sourceID, currentSourceObject.source.name)}
  >
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    import local file to refresh source
  </MenuItem>
{:else}
  <MenuItem
    icon
    on:select={() => onRefreshSource(sourceID, currentSourceObject.source.name)}
  >
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
    dispatch("rename-asset");
  }}
>
  <EditIcon slot="icon" />

  rename...
</MenuItem>
<!-- FIXME: this should pop up an "are you sure?" modal -->
<MenuItem
  icon
  on:select={() => handleDeleteSource(currentSourceObject.source.name)}
>
  <Cancel slot="icon" />
  delete</MenuItem
>
