<script lang="ts">
  import {
    getRuntimeServiceGetCatalogEntryQueryKey,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceTriggerRefresh,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import {
    useSourceFromYaml,
    useSourceNames,
  } from "@rilldata/web-local/lib/svelte-query/sources";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtimeStore } from "../../../application-state-stores/application-store";
  import { overlay } from "../../../application-state-stores/overlay-store";
  import { navigationEvent } from "../../../metrics/initMetrics";
  import { deleteFileArtifact } from "../../../svelte-query/actions";
  import { useModelNames } from "../../../svelte-query/models";
  import { getFileFromName } from "../../../util/entity-mappers";
  import Cancel from "../../icons/Cancel.svelte";
  import EditIcon from "../../icons/EditIcon.svelte";
  import Explore from "../../icons/Explore.svelte";
  import Import from "../../icons/Import.svelte";
  import Model from "../../icons/Model.svelte";
  import RefreshIcon from "../../icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "../../menu";
  import { createModelFromSource } from "../models/createModel";
  import { refreshSource } from "./refreshSource";

  export let sourceName: string;

  const queryClient = useQueryClient();

  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  $: sourceNames = useSourceNames($runtimeStore.instanceId);
  $: sourceFromYaml = useSourceFromYaml(
    $runtimeStore.instanceId,
    getFileFromName(sourceName, EntityType.Table)
  );
  $: getSource = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  let source: V1Source;
  $: source = $getSource?.data?.entry?.source;

  const dispatch = createEventDispatcher();

  $: runtimeInstanceId = $runtimeStore.instanceId;
  $: getSource = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );

  const deleteSource = useRuntimeServiceDeleteFileAndReconcile();
  const refreshSourceMutation = useRuntimeServiceTriggerRefresh();
  const createEntityMutation = useRuntimeServicePutFileAndReconcile();
  $: modelNames = useModelNames($runtimeStore.instanceId);

  const handleDeleteSource = async (tableName: string) => {
    await deleteFileArtifact(
      queryClient,
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $deleteSource,
      $appStore.activeEntity,
      $sourceNames.data
    );
    toggleMenu();
  };

  const handleCreateModel = async (tableName: string) => {
    try {
      const previousActiveEntity = $appStore.activeEntity?.type;
      const newModelName = await createModelFromSource(
        queryClient,
        runtimeInstanceId,
        $modelNames.data,
        tableName,
        $createEntityMutation
      );

      navigationEvent.fireEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        EntityTypeToScreenMap[previousActiveEntity],
        MetricsEventScreenName.Model
      );
    } catch (err) {
      console.error(err);
    }
  };

  const bootstrapDashboard = async (_tableName: string) => {
    // const previousActiveEntity = $rillAppStore?.activeEntity?.type;
    // const createdMetricsId = await autoCreateMetricsDefinitionForSource(
    //   $persistentModelStore.entities,
    //   $derivedTableStore.entities,
    //   sourceID,
    //   tableName
    // );
    //
    // navigationEvent.fireEvent(
    //   createdMetricsId,
    //   BehaviourEventMedium.Menu,
    //   MetricsEventSpace.LeftPanel,
    //   EntityTypeToScreenMap[previousActiveEntity],
    //   MetricsEventScreenName.Dashboard
    // );
  };

  const onRefreshSource = async (tableName: string) => {
    const connector: string =
      $getSource?.data?.entry.source?.connector ?? $sourceFromYaml.data?.type;
    if (!connector) {
      // if parse failed or there is no catalog entry, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }

    try {
      await refreshSource(
        connector,
        tableName,
        runtimeInstanceId,
        $refreshSourceMutation,
        $createEntityMutation
      );

      // invalidate the data preview (async)
      // TODO: use new runtime approach
      // Old approach: dataModelerService.dispatch("collectTableInfo", [currentSource?.id]);

      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
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
  disabled={!schemaHasTimestampColumn(source?.schema)}
  icon
  on:select={() => bootstrapDashboard(sourceName)}
>
  <Explore slot="icon" />
  autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !schemaHasTimestampColumn(source?.schema)}
      requires a timestamp column
    {/if}
  </svelte:fragment>
</MenuItem>

{#if $getSource?.data?.entry?.source?.connector === "file"}
  <MenuItem icon on:select={() => onRefreshSource(sourceName)}>
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    import local file to refresh source
  </MenuItem>
{:else}
  <MenuItem icon on:select={() => onRefreshSource(sourceName)}>
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
