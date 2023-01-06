<script lang="ts">
  import { goto } from "$app/navigation";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import {
    useSourceFromYaml,
    useSourceNames,
  } from "@rilldata/web-common/features/sources/selectors";
  import {
    getRuntimeServiceGetCatalogEntryQueryKey,
    useRuntimeServiceDeleteFileAndReconcile,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceRefreshAndReconcile,
    V1ReconcileResponse,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import { navigationEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    EntityTypeToScreenMap,
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import {
    deleteFileArtifact,
    useCreateDashboardFromSource,
  } from "@rilldata/web-local/lib/svelte-query/actions";
  import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";
  import { useDashboardNames } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { getName } from "@rilldata/web-local/lib/util/incrementName";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { EntityType } from "../../../lib/entity";
  import { useModelNames } from "../../models/selectors";
  import { createModelFromSource } from "../createModel";
  import { refreshSource } from "../refreshSource";

  export let sourceName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtimeStore.instanceId;

  const dispatch = createEventDispatcher();

  $: getSource = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  let source: V1Source;
  $: source = $getSource?.data?.entry?.source;
  $: embedded = $getSource?.data?.entry?.embedded;
  $: path = source?.properties?.path;

  $: sourceFromYaml = useSourceFromYaml(
    $runtimeStore.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  $: sourceNames = useSourceNames($runtimeStore.instanceId);
  $: modelNames = useModelNames($runtimeStore.instanceId);
  $: dashboardNames = useDashboardNames($runtimeStore.instanceId);

  const deleteSource = useRuntimeServiceDeleteFileAndReconcile();
  const refreshSourceMutation = useRuntimeServiceRefreshAndReconcile();
  const createEntityMutation = useRuntimeServicePutFileAndReconcile();
  const createDashboardFromSourceMutation = useCreateDashboardFromSource();
  const createFileMutation = useRuntimeServicePutFileAndReconcile();

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

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appStore.activeEntity?.type;
      const newModelName = await createModelFromSource(
        queryClient,
        runtimeInstanceId,
        $modelNames.data,
        sourceName,
        embedded ? `"${path}"` : sourceName,
        $createFileMutation
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
          const previousActiveEntity = $appStore?.activeEntity?.type;
          navigationEvent.fireEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            EntityTypeToScreenMap[previousActiveEntity],
            MetricsEventScreenName.Dashboard
          );
          return invalidateAfterReconcile(queryClient, runtimeInstanceId, resp);
        },
        onSettled: () => {
          overlay.set(null);
          toggleMenu(); // unmount component
        },
      }
    );
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
        $createEntityMutation,
        queryClient
      );

      // invalidate the data preview (async)
      // TODO: use new runtime approach
      // Old approach: dataModelerService.dispatch("collectTableInfo", [currentSource?.id]);

      // invalidate the "refreshed_on" time
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
        runtimeInstanceId,
        tableName
      );
      await queryClient.refetchQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };
</script>

<MenuItem icon on:select={() => handleCreateModel(sourceName)}>
  <Model slot="icon" />
  Create new model
</MenuItem>

<MenuItem
  disabled={!schemaHasTimestampColumn(source?.schema)}
  icon
  on:select={() => handleCreateDashboardFromSource(sourceName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if !schemaHasTimestampColumn(source?.schema)}
      Requires a timestamp column
    {/if}
  </svelte:fragment>
</MenuItem>

{#if $getSource?.data?.entry?.source?.connector === "local_file"}
  <MenuItem icon on:select={() => onRefreshSource(sourceName)}>
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    Import local file to refresh source
  </MenuItem>
{:else}
  <MenuItem icon on:select={() => onRefreshSource(sourceName)}>
    <svelte:fragment slot="icon">
      <RefreshIcon />
    </svelte:fragment>
    Refresh source data
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
  Rename...
</MenuItem>
<!-- FIXME: this should pop up an "are you sure?" modal -->
<MenuItem
  icon
  on:select={() => handleDeleteSource(sourceName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  Delete
</MenuItem>
