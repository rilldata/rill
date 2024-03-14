<script lang="ts">
  import { page } from "$app/stores";
  import DashboardEmbed from "@rilldata/web-admin/features/embeds/DashboardEmbed.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  // Get the query parameters from the iframe's URL
  const instanceId = $page.url.searchParams.get("instance_id");
  const resourceName = $page.url.searchParams.get("resource");
  const resourceKind = $page.url.searchParams.get("kind");
  const navigation = $page.url.searchParams.get("navigation");
  // Ignoring state and theme params for now
  // const state = $page.url.searchParams.get("state");
  // const theme = $page.url.searchParams.get("theme");
</script>

<svelte:head>
  <title>{resourceName} - Rill</title>
</svelte:head>

{#if resourceKind === ResourceKind.MetricsView.toString()}
  <DashboardEmbed {instanceId} dashboardName={resourceName} />
{/if}
