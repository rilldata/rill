<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    getRuntimeServiceGetCatalogObjectQueryKey,
    getRuntimeServiceListCatalogObjectsQueryKey,
    getRuntimeServiceListFilesQueryKey,
    RuntimeServiceListCatalogObjectsType,
    useRuntimeServiceDeleteFileAndMigrate,
    useRuntimeServiceGetCatalogObject,
    useRuntimeServicePutFileAndMigrate,
    useRuntimeServiceTriggerRefresh,
  } from "@rilldata/web-common/runtime-client";
  import { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/common/metrics-service/MetricsTypes";
  import { getName } from "@rilldata/web-local/common/utils/incrementName";
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
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import {
    useSourceFromYaml,
    useSourceNames,
  } from "@rilldata/web-local/lib/svelte-query/sources";
  import { createEventDispatcher, getContext } from "svelte";
  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { getNextEntityName } from "../../../../common/utils/getNextEntityId";
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

  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  $: sourceNames = useSourceNames($runtimeStore.repoId);
  $: sourceFromYaml = useSourceFromYaml(
    $runtimeStore.repoId,
    `/sources/${sourceName}.yaml`
  );

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
    sourceName
  );

  $: sourceID = $persistentTableStore.entities.find(
    (entity) => entity.tableName === sourceName
  );
  $: derivedTable = $derivedTableStore?.entities?.find(
    (source) => source.id === sourceID
  );

  const deleteSource = useRuntimeServiceDeleteFileAndMigrate();
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createSource = useRuntimeServicePutFileAndMigrate();
  $: modelNames = useModelNames($runtimeStore.repoId);
  const createModel = useRuntimeServicePutFileAndMigrate();

  const handleDeleteSource = async (tableName: string) => {
    try {
      await $deleteSource.mutateAsync({
        data: {
          repoId: $runtimeStore.repoId,
          instanceId: runtimeInstanceId,
          path: `sources/${tableName}.yaml`,
        },
      });
      if (
        $rillAppStore.activeEntity.type === EntityType.Table &&
        $rillAppStore.activeEntity.id === sourceName
      ) {
        const nextSourceName = getNextEntityName($sourceNames.data, sourceName);
        if (nextSourceName) {
          goto(`/source/${nextSourceName}`);
        } else {
          goto("/");
        }
      }
      sourceUpdated(tableName);
      await queryClient.invalidateQueries(
        getRuntimeServiceListFilesQueryKey($runtimeStore.repoId)
      );
    } catch (err) {
      console.error(err);
    }
    toggleMenu();
  };

  const handleCreateModel = (tableName: string) => {
    const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    const newModelName = getName(`${tableName}_model`, modelNames);

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
    const connector: string =
      $getSource?.data?.object?.source?.connector ?? $sourceFromYaml.data?.type;
    if (!connector) {
      // if parse failed or there is no catalog object, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }

    try {
      await refreshSource(
        connector,
        tableName,
        $runtimeStore,
        $refreshSourceMutation,
        $createSource
      );

      if (id) {
        // invalidate the data preview (async)
        dataModelerService.dispatch("collectTableInfo", [id]);
      }

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

{#if $getSource?.data?.object?.source?.connector === "file"}
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
  on:select={() => handleDeleteSource(sourceName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  delete</MenuItem
>
