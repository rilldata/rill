<script lang="ts">
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import {
    useIsLocalFileConnector,
    useSourceFromYaml,
  } from "@rilldata/web-common/features/sources/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
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
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createModelFromSource } from "../createModel";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "../refreshSource";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  const dispatch = createEventDispatcher();

  const { customDashboards } = featureFlags;

  $: sourceQuery = fileArtifact.getResource(queryClient, runtimeInstanceId);
  let source: V1SourceV2 | undefined;
  $: source = $sourceQuery.data?.source;
  $: sinkConnector = $sourceQuery.data?.source?.spec?.sinkConnector;
  $: sourceHasError = fileArtifact.getHasErrors(queryClient, runtimeInstanceId);
  $: sourceIsIdle =
    $sourceQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $sourceHasError || !sourceIsIdle;
  $: tableName = source?.state?.table ?? "";

  $: sourceFromYaml = useSourceFromYaml($runtime.instanceId, filePath);

  $: modelNames = useModelFileNames($runtime.instanceId);

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    sinkConnector as string,
    "",
    "",
    tableName,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appScreen?.type;
      const newModelName = await createModelFromSource(
        runtimeInstanceId,
        $modelNames.data ?? [],
        tableName,
        tableName,
      );

      await behaviourEvent.fireNavigationEvent(
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

  const onRefreshSource = async () => {
    const connector: string | undefined =
      source?.state?.connector ?? $sourceFromYaml.data?.type;
    if (!connector) {
      // if parse failed or there is no catalog entry, we cannot refresh source
      // TODO: show the import source modal with fixed tableName
      return;
    }
    try {
      await refreshSource(
        connector,
        filePath,
        $sourceQuery.data?.meta?.name?.name ?? "",
        runtimeInstanceId,
      );
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(
    $runtime.instanceId,
    filePath,
  );
  $: isLocalFileConnector = $isLocalFileConnectorQuery.data;

  async function onReplaceSource() {
    await replaceSourceWithUploadedFile(runtimeInstanceId, filePath);
    overlay.set(null);
  }
</script>

<NavigationMenuItem on:click={handleCreateModel}>
  <Model slot="icon" />
  Create new model
</NavigationMenuItem>

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createDashboardFromTable}
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
</NavigationMenuItem>
{#if $customDashboards}
  <NavigationMenuItem
    disabled={disableCreateDashboard}
    on:click={() => {
      dispatch("generate-chart", {
        table: source?.state?.table,
        connector: source?.state?.connector,
      });
    }}
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
  </NavigationMenuItem>
{/if}

<NavigationMenuItem on:click={onRefreshSource}>
  <RefreshIcon slot="icon" />
  Refresh source
</NavigationMenuItem>

{#if isLocalFileConnector}
  <NavigationMenuItem on:click={onReplaceSource}>
    <Import slot="icon" />
    Replace source with uploaded file
  </NavigationMenuItem>
{/if}
