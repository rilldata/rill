<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import {
    useIsLocalFileConnector,
    useSource,
    useSourceFileNames,
    useSourceFromYaml,
  } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { EntityType } from "../../entity-management/types";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createModelFromSource } from "../createModel";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "../refreshSource";

  export let sourceName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;
  $: filePath = getFilePathFromNameAndType(sourceName, EntityType.Table);

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  const dispatch = createEventDispatcher();

  const { customDashboards } = featureFlags;

  $: sourceQuery = useSource(runtimeInstanceId, sourceName);
  let source: V1SourceV2 | undefined;
  $: source = $sourceQuery.data?.source;
  $: embedded = false; // TODO: remove embedded support
  $: path = source?.spec?.properties?.path;
  $: sourceHasError = getFileHasErrors(
    queryClient,
    runtimeInstanceId,
    filePath,
  );
  $: sourceIsIdle =
    $sourceQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $sourceHasError || !sourceIsIdle;

  $: sourceFromYaml = useSourceFromYaml($runtime.instanceId, filePath);

  $: sourceNames = useSourceFileNames($runtime.instanceId);
  $: modelNames = useModelFileNames($runtime.instanceId);

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    sourceName,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  const handleDeleteSource = async (tableName: string) => {
    await deleteFileArtifact(
      runtimeInstanceId,
      tableName,
      EntityType.Table,
      $sourceNames.data ?? [],
    );
    toggleMenu();
  };

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appScreen?.type;
      const newModelName = await createModelFromSource(
        runtimeInstanceId,
        $modelNames.data ?? [],
        sourceName,
        embedded ? `"${path}"` : sourceName,
      );

      behaviourEvent.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousActiveEntity,
        MetricsEventScreenName.Model,
      );
    } catch (err) {
      console.error(err);
    }
  };

  const onRefreshSource = async (tableName: string) => {
    const connector: string | undefined =
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

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(
    $runtime.instanceId,
    sourceName,
  );
  $: isLocalFileConnector = $isLocalFileConnectorQuery.data;

  async function onReplaceSource(sourceName: string) {
    await replaceSourceWithUploadedFile(runtimeInstanceId, sourceName);
    overlay.set(null);
  }
</script>

<MenuItem icon on:select={() => handleCreateModel()}>
  <Model slot="icon" />
  Create new model
</MenuItem>

<MenuItem
  disabled={disableCreateDashboard}
  icon
  on:select={() => {
    void createDashboardFromTable();
    toggleMenu();
  }}
  propogateSelect={false}
>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard with AI
    <WandIcon class="w-3 h-3" />
  </div>
  <svelte:fragment slot="description">
    {#if $sourceHasError}
      Source has errors
    {:else if !sourceIsIdle}
      Source is being ingested
    {/if}
  </svelte:fragment>
</MenuItem>
{#if customDashboards}
  <MenuItem
    disabled={disableCreateDashboard}
    icon
    on:select={() => {
      dispatch("generate-chart", {
        table: source?.state?.table,
        connector: source?.state?.connector,
      });
      toggleMenu();
    }}
    propogateSelect={false}
  >
    <Explore slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate chart with AI
      <WandIcon class="w-3 h-3" />
    </div>
    <svelte:fragment slot="description">
      {#if $sourceHasError}
        Source has errors
      {:else if !sourceIsIdle}
        Source is being ingested
      {/if}
    </svelte:fragment>
  </MenuItem>
{/if}

<MenuItem icon on:select={() => onRefreshSource(sourceName)}>
  <svelte:fragment slot="icon">
    <RefreshIcon />
  </svelte:fragment>
  Refresh source
</MenuItem>

{#if isLocalFileConnector}
  <MenuItem icon on:select={() => onReplaceSource(sourceName)}>
    <svelte:fragment slot="icon">
      <Import />
    </svelte:fragment>
    Replace source with uploaded file
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
