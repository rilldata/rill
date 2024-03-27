<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import DashboardEmbed from "@rilldata/web-admin/features/embeds/DashboardEmbed.svelte";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  const instanceId = $page.url.searchParams.get("instance_id");
  const initialResourceName = $page.url.searchParams.get("resource");
  const resourceKind = $page.url.searchParams.get("kind");
  const navigation = $page.url.searchParams.get("navigation");
  // Ignoring state and theme params for now
  // const state = $page.url.searchParams.get("state");
  // const theme = $page.url.searchParams.get("theme");

  // Manage active resource
  let activeResourceName = initialResourceName;

  function handleSelectDashboard(event: CustomEvent<string>) {
    activeResourceName = event.detail;
  }

  function handleGoHome() {
    activeResourceName = "";
  }
</script>

<svelte:head>
  <title>{activeResourceName} - Rill</title>
</svelte:head>

{#if navigation}
  <TopNavigationBarEmbed
    {instanceId}
    {activeResourceName}
    on:select-dashboard={handleSelectDashboard}
    on:go-home={handleGoHome}
  />
{/if}
<!-- Metrics Explorers -->
{#if resourceKind === ResourceKind.MetricsView.toString()}
  {#if !activeResourceName}
    <ContentContainer>
      <div class="flex flex-col items-center">
        <DashboardsTable
          isEmbedded
          on:select-dashboard={handleSelectDashboard}
        />
      </div>
    </ContentContainer>
  {:else}
    <DashboardEmbed {instanceId} dashboardName={activeResourceName} />
  {/if}
{/if}
