<script lang="ts">
  import { goto } from "$app/navigation";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import {
    useSource,
    useSourceFileNames,
    useSourceFromYaml,
  } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { getName } from "../../entity-management/name-utils";
  import { EntityType } from "../../entity-management/types";
  import { useCreateDashboardFromSource } from "../createDashboard";
  import { createModelFromSource } from "../createModel";
  import { refreshSource } from "../refreshSource";

  export let sourceName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;
  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  const dispatch = createEventDispatcher();

  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2;
  $: source = $sourceQuery.data?.source;
  $: embedded = false; // TODO: remove embedded support
  $: path = source?.spec?.properties?.path;
  $: sourceHasError = getFileHasErrors(
    queryClient,
    runtimeInstanceId,
    filePath
  );
  $: sourceIsIdle =
    $sourceQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $sourceHasError || !sourceIsIdle;

  $: sourceFromYaml = useSourceFromYaml($runtime.instanceId, filePath);

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: modelNames = useModelFileNames($runtime.instanceId);
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);

  const createDashboardFromSourceMutation = useCreateDashboardFromSource();

  const handleDeleteSource = async (tableName: string) => {
    await deleteFileArtifact(
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $sourceNames.data
    );
    toggleMenu();
  };

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appScreen?.type;
      const newModelName = await createModelFromSource(
        runtimeInstanceId,
        $modelNames.data,
        sourceName,
        embedded ? `"${path}"` : sourceName
      );

      behaviourEvent.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousActiveEntity,
        MetricsEventScreenName.Model
      );
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateDashboardFromSource = async (sourceName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + sourceName,
    });
    const newModelName = getName(`${sourceName}_model`, $modelNames.data);
    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      $dashboardNames.data
    );
    await waitUntil(() => !!$sourceQuery.data);
    $createDashboardFromSourceMutation.mutate(
      {
        data: {
          instanceId: $runtime.instanceId,
          sourceResource: $sourceQuery.data,
          newModelName,
          newDashboardName,
        },
      },
      {
        onSuccess: async () => {
          goto(`/dashboard/${newDashboardName}`);
          const previousActiveEntity = $appScreen?.type;
          behaviourEvent.fireNavigationEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            previousActiveEntity,
            MetricsEventScreenName.Dashboard
          );
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
      source?.state?.connector ?? $sourceFromYaml.data?.type;
    if (!connector) {
      // if parse failed or there is no catalog entry, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }
    try {
      await refreshSource(connector, tableName, runtimeInstanceId);
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
  disabled={disableCreateDashboard}
  icon
  on:select={() => handleCreateDashboardFromSource(sourceName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if $sourceHasError}
      Source has errors
    {:else if !sourceIsIdle}
      Source is being ingested
    {/if}
  </svelte:fragment>
</MenuItem>

{#if source?.state?.connector === "local_file"}
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
