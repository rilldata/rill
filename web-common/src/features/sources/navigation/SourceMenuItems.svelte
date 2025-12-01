<script lang="ts">
  import { goto } from "$app/navigation";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import Import from "@rilldata/web-common/components/icons/Import.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import {
    useIsLocalFileConnector,
    useSourceFromYaml,
  } from "@rilldata/web-common/features/sources/selectors";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1Source } from "@rilldata/web-common/runtime-client";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "../refreshSource";
  import { createSqlModelFromTable } from "../../connectors/code-utils";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  $: ({ instanceId } = $runtime);

  const { ai } = featureFlags;

  $: sourceQuery = fileArtifact.getResource(queryClient, instanceId);
  let source: V1Source | undefined;
  $: source = $sourceQuery.data?.source;
  $: sinkConnector = $sourceQuery.data?.source?.spec?.sinkConnector;
  $: sourceHasError = fileArtifact.getHasErrors(queryClient, instanceId);
  $: sourceIsIdle =
    $sourceQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $sourceHasError || !sourceIsIdle;
  $: connector = source?.state?.connector as string;
  const database = ""; // Sources are ingested into the default database
  const databaseSchema = ""; // Sources are ingested into the default database schema
  $: tableName = source?.state?.table as string;

  $: sourceFromYaml = useSourceFromYaml(instanceId, filePath);

  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    sinkConnector as string,
    database,
    databaseSchema,
    tableName,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  $: createExploreFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    sinkConnector as string,
    database,
    databaseSchema,
    tableName,
    true,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = getScreenNameFromPage();
      const addDevLimit = false; // Typically, the `dev` limit would be applied on the Source itself
      const [newModelPath, newModelName] = await createSqlModelFromTable(
        queryClient,
        connector,
        database,
        databaseSchema,
        tableName,
        addDevLimit,
      );
      await goto(`/files${newModelPath}`);
      await behaviourEvent?.fireNavigationEvent(
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
        instanceId,
      );
    } catch {
      // no-op
    }
  };

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  $: isLocalFileConnector = $isLocalFileConnectorQuery.data;

  async function onReplaceSource() {
    await replaceSourceWithUploadedFile(instanceId, filePath);
    overlay.set(null);
  }
</script>

<NavigationMenuItem on:click={handleCreateModel}>
  <Model slot="icon" />
  Create new model
</NavigationMenuItem>

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createMetricsViewFromTable}
>
  <MetricsViewIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate metrics
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
  <svelte:fragment slot="description">
    {#if $sourceHasError}
      Source has errors
    {:else if !sourceIsIdle}
      Source is being ingested
    {/if}
  </svelte:fragment>
</NavigationMenuItem>
<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createExploreFromTable}
>
  <ExploreIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
  <svelte:fragment slot="description">
    {#if $sourceHasError}
      Source has errors
    {:else if !sourceIsIdle}
      Source is being ingested
    {/if}
  </svelte:fragment>
</NavigationMenuItem>

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

<NavigationMenuSeparator />
