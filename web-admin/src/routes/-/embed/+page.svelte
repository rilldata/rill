<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import CustomDashboardEmbed from "@rilldata/web-admin/features/embeds/CustomDashboardEmbed.svelte";
  import MetricsExplorerEmbed from "@rilldata/web-admin/features/embeds/MetricsExplorerEmbed.svelte";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1ResourceName } from "@rilldata/web-common/runtime-client";

  const instanceId = $page.url.searchParams.get("instance_id");
  const initialResourceName = $page.url.searchParams.get("resource");
  const initialResourceKind = $page.url.searchParams.get("kind");
  const navigation = $page.url.searchParams.get("navigation");
  // Ignoring state and theme params for now
  // const state = $page.url.searchParams.get("state");
  // const theme = $page.url.searchParams.get("theme");

  // Manage active resource
  let activeResource: V1ResourceName | null = null;
  if (initialResourceName && initialResourceKind) {
    activeResource = {
      name: initialResourceName,
      kind: initialResourceKind,
    };
  }

  function handleSelectResource(event: CustomEvent<V1ResourceName>) {
    activeResource = event.detail;
  }

  function handleGoHome() {
    activeResource = null;
  }
</script>

<svelte:head>
  {#if activeResource}
    <title>{activeResource.name} - Rill</title>
  {:else}
    <title>Rill</title>
  {/if}
</svelte:head>

{#if navigation}
  <TopNavigationBarEmbed
    {instanceId}
    {activeResource}
    on:select-resource={handleSelectResource}
    on:go-home={handleGoHome}
  />

  {#if !activeResource}
    <ContentContainer>
      <div class="flex flex-col items-center">
        <DashboardsTable isEmbedded on:select-resource={handleSelectResource} />
      </div>
    </ContentContainer>
  {/if}
{/if}

{#if activeResource?.kind === ResourceKind.MetricsView.toString()}
  <MetricsExplorerEmbed {instanceId} dashboardName={activeResource.name} />
{:else if activeResource?.kind === ResourceKind.Dashboard.toString()}
  <CustomDashboardEmbed {instanceId} dashboardName={activeResource.name} />
{/if}
