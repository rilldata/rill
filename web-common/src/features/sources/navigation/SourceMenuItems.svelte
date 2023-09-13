<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { deleteFile } from "@rilldata/web-common/features/entity-management/file-actions";
  import { useAllEntityNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createDashboardFromModel } from "@rilldata/web-common/features/models/createDashboardFromModel";
  import {
    useSourceFromYaml,
    useSourceNames,
  } from "@rilldata/web-common/features/sources/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { getLeftPanelParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceRefreshAndReconcile,
    getRuntimeServiceGetCatalogEntryQueryKey,
    V1Source,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createModelFromSourceCreator } from "web-common/src/features/sources/createModelFromSource";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EntityType } from "../../entity-management/types";
  import { refreshSource } from "../refreshSource";

  export let sourceName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  const dispatch = createEventDispatcher();

  $: getSource = createRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    sourceName
  );
  let source: V1Source;
  $: source = $getSource?.data?.entry?.source;
  $: embedded = $getSource?.data?.entry?.embedded;
  $: path = source?.properties?.path;
  $: hasNoSourceCatalog = !source;

  $: sourceFromYaml = useSourceFromYaml(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  $: sourceNames = useSourceNames($runtime.instanceId);
  $: allEntityNames = useAllEntityNames($runtime.instanceId);

  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();
  const createEntityMutation = createRuntimeServicePutFileAndReconcile();
  const dashboardFromSourceCreator = createModelFromSourceCreator(
    allEntityNames,
    undefined,
    createDashboardFromModel(allEntityNames, getLeftPanelParams())
  );

  $: modelFromSourceCreator = createModelFromSourceCreator(
    allEntityNames,
    getLeftPanelParams()
  );

  const handleDeleteSource = async (tableName: string) => {
    await deleteFile(
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $sourceNames.data
    );
    toggleMenu();
  };

  const handleCreateModel = async () => {
    try {
      await modelFromSourceCreator(
        undefined, // TODO
        embedded ? `"${path}"` : sourceName,
        "/models/"
      );
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateDashboardFromSource = async (sourceName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });

    await dashboardFromSourceCreator(undefined, sourceName);

    // TODO: should this wait till everything is finished?
    overlay.set(null);
    toggleMenu(); // unmount component
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
        queryClient,
        connector === "s3" || connector === "gcs" || connector === "https"
          ? source?.properties?.path
          : sourceName
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

<MenuItem icon on:select={() => handleCreateModel()}>
  <Model slot="icon" />
  Create new model
</MenuItem>

<MenuItem
  disabled={hasNoSourceCatalog}
  icon
  on:select={() => handleCreateDashboardFromSource(sourceName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if hasNoSourceCatalog}
      Source has errors
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
