<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createFileDeleter } from "@rilldata/web-common/features/entity-management/file-actions";
  import { createEntityRefresher } from "@rilldata/web-common/features/entity-management/refresh-entity";
  import {
    useAllEntityNames,
    useSource,
    useSourceNames,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createDashboardFromModelCreator } from "@rilldata/web-common/features/models/createDashboardFromModel";
  import { useSourceFromYaml } from "@rilldata/web-common/features/sources/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { getLeftPanelParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
  import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createModelFromSourceCreator } from "web-common/src/features/sources/createModelFromSource";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EntityType } from "../../entity-management/types";

  export let sourceName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  $: runtimeInstanceId = $runtime.instanceId;

  const dispatch = createEventDispatcher();

  $: getSource = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2;
  $: source = $getSource?.data?.source;
  $: embedded = false; // TODO
  $: path = source?.spec?.properties?.path;
  $: hasError = !!$getSource?.data?.meta.reconcileError;

  $: sourceFromYaml = useSourceFromYaml(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );

  $: sourceNames = useSourceNames($runtime.instanceId);
  $: allEntityNames = useAllEntityNames($runtime.instanceId);

  $: dashboardFromSourceCreator = createModelFromSourceCreator(
    allEntityNames,
    undefined,
    createDashboardFromModelCreator(allEntityNames, getLeftPanelParams())
  );
  $: fileDeleter = createFileDeleter(sourceNames);
  const sourceRefresher = createEntityRefresher();

  $: modelFromSourceCreator = createModelFromSourceCreator(
    allEntityNames,
    getLeftPanelParams()
  );

  const handleDeleteSource = async (tableName: string) => {
    await fileDeleter(tableName, EntityType.Table);
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

  const onRefreshSource = async () => {
    const connector: string =
      source?.state?.connector ?? $sourceFromYaml.data?.type;
    if (!connector) {
      // if parse failed or there is no catalog entry, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }
    try {
      await sourceRefresher($getSource.data);
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
  disabled={hasError}
  icon
  on:select={() => handleCreateDashboardFromSource(sourceName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if hasError}
      Source has errors
    {/if}
  </svelte:fragment>
</MenuItem>

{#if source?.state?.connector === "local_file"}
  <MenuItem icon on:select={onRefreshSource}>
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    Import local file to refresh source
  </MenuItem>
{:else}
  <MenuItem icon on:select={onRefreshSource}>
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
