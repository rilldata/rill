<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    getRuntimeServiceListCatalogObjectsQueryKey,
    getRuntimeServiceListFilesQueryKey,
    RuntimeServiceListCatalogObjectsType,
    useRuntimeServiceDeleteFileAndMigrate,
    useRuntimeServiceGetCatalogObject,
    useRuntimeServiceListCatalogObjects,
    useRuntimeServicePutFileAndMigrate,
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
  import type { ApplicationStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import type {
    DerivedTableStore,
    PersistentTableStore,
  } from "@rilldata/web-local/lib/application-state-stores/table-stores";
  import {
    autoCreateMetricsDefinitionForSource,
    sourceUpdated,
  } from "@rilldata/web-local/lib/redux-store/source/source-apis";
  import { derivedProfileEntityHasTimestampColumn } from "@rilldata/web-local/lib/redux-store/source/source-selectors";
  import { createEventDispatcher, getContext } from "svelte";
  import { getName } from "../../../../common/utils/incrementName";
  import {
    dataModelerService,
    runtimeStore,
  } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { queryClient } from "../../../svelte-query/globalQueryClient";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import Import from "../../icons/Import.svelte";
  import Model from "../../icons/Model.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "../../menu";
  import { refreshSource } from "./refreshSource";

  export let sourceName: string;
  export let sourceID: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

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

  $: runtimeInstanceId = $runtimeStore.instanceId;
  $: getSource = useRuntimeServiceGetCatalogObject(
    runtimeInstanceId,
    persistentTable.tableName
  );

  const deleteSource = useRuntimeServiceDeleteFileAndMigrate();
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  $: getModels = useRuntimeServiceListCatalogObjects(runtimeInstanceId, {
    type: RuntimeServiceListCatalogObjectsType.TYPE_MODEL,
  });
  const createModel = useRuntimeServicePutFileAndMigrate();
  const createSource = useRuntimeServicePutFileAndMigrate();

  $: persistentTable = $persistentTableStore?.entities?.find(
    (source) => source.id === sourceID
  );

  $: derivedTable = $derivedTableStore?.entities?.find(
    (source) => source.id === sourceID
  );

  const handleDeleteSource = (tableName: string) => {
    $deleteSource.mutate(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: runtimeInstanceId,
          path: `sources/${tableName}.yaml`,
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
              sourceID
            );
            const nextSourceName = $persistentTableStore.entities.find(
              (source) => source.id === nextSourceId
            ).tableName;
            if (nextSourceName) {
              goto(`/source/${nextSourceName}`);
            } else {
              goto("/");
            }
          }
          sourceUpdated(tableName);
          return queryClient.invalidateQueries(
            getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
          );
        },
        onError: (error) => {
          console.error(error);
        },
        onSettled: () => {
          // onSettled gets triggered *after* both onSuccess and onError
          toggleMenu();
        },
      }
    );
  };

  const handleCreateModel = (tableName: string) => {
    const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    const newModelName = getName(
      `${tableName}_model`,
      $getModels.data.objects.map((object) => object.name)
    );

    $createModel.mutateAsync(
      {
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: $runtimeStore.instanceId,
          path: `models/${newModelName}.sql`,
          blob: `select * from ${tableName}`,
          create: true,
          createOnly: true,
          strict: true,
        },
      },
      {
        onSuccess: (res) => {
          if (res.errors) {
            res.errors.forEach((error) => {
              console.error(error);
            });
            return;
          }

          navigationEvent.fireEvent(
            newModelName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            EntityTypeToScreenMap[previousActiveEntity],
            MetricsEventScreenName.Model
          );

          queryClient.invalidateQueries(
            getRuntimeServiceListCatalogObjectsQueryKey(
              $runtimeStore.instanceId,
              {
                type: RuntimeServiceListCatalogObjectsType.TYPE_MODEL,
              }
            )
          );
        },
      }
    );
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
        $getSource.data.object.source.connector,
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

<MenuItem icon on:select={() => handleCreateModel(sourceName)}>
  <Model slot="icon" />
  create new model
</MenuItem>

<MenuItem
  disabled={!derivedProfileEntityHasTimestampColumn(derivedTable)}
  icon
  on:select={() => bootstrapDashboard(sourceID, sourceName)}
>
  <Explore slot="icon" />
  autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !derivedProfileEntityHasTimestampColumn(derivedTable)}
      requires a timestamp column
    {/if}
  </svelte:fragment>
</MenuItem>

{#if $getSource.data.object.source.connector === "file"}
  <MenuItem icon on:select={() => onRefreshSource(sourceID, sourceName)}>
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    import local file to refresh source
  </MenuItem>
{:else}
  <MenuItem icon on:select={() => onRefreshSource(sourceID, sourceName)}>
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
  propogateSelect={false}
  on:select={() => handleDeleteSource(sourceName)}
>
  <Cancel slot="icon" />
  delete</MenuItem
>
