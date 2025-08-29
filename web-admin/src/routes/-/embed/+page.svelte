<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import CanvasEmbed from "@rilldata/web-admin/features/embeds/CanvasEmbed.svelte";
  import ExploreEmbed, {
    EmbedStorageNamespacePrefix,
  } from "@rilldata/web-admin/features/embeds/ExploreEmbed.svelte";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import UnsupportedKind from "@rilldata/web-admin/features/embeds/UnsupportedKind.svelte";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { V1ResourceName } from "@rilldata/web-common/runtime-client";

  // Embedded dashboards communicate directly with the project runtime and do not communicate with the admin server.
  // One by-product of this is that they have no access to control plane features like alerts, bookmarks, and scheduled reports.
  featureFlags.set(false, "adminServer");

  const instanceId = $page.url.searchParams.get("instance_id");
  $: resourceName = $page.url.searchParams.get("resource");
  $: resourceType =
    $page.url.searchParams.get("type") ?? $page.url.searchParams.get("kind"); // "kind" is for backwards compatibility
  const navigation = $page.url.searchParams.get("navigation");
  const navigationEnabled = navigation === "true";
  // Ignoring state and theme params for now
  // const state = $page.url.searchParams.get("state");
  // const theme = $page.url.searchParams.get("theme");

  // Manage active resource
  let activeResource: V1ResourceName | null = null;
  $: if (resourceName && resourceType) {
    cleanSessionStorageForResource(resourceType, resourceName);
    activeResource = {
      name: resourceName,
      kind: resourceType,
    };
  } else {
    // Important! Do not set `activeResource` to `null` here.
    // In `DashboardStateDataLoader` we are currently clearing the url of non-dashboard params since it will interfere with session storage extraction logic.
    // So setting `activeResource` to `null` will make the navigation logic in this file to think there is no active resource everytime `DashboardStateDataLoader` cleaned the url.
    // For going to home we manually set `activeResource` to `null` in `handleGoHome`.
    // TODO: move to a route based approach to avoid this issues.
  }

  const resourcesSeen = new Set<string>();
  // Clean session storage for dashboards that are navigated to for the 1st time.
  // This way once the page is loaded, the dashboard state is persisted.
  // But the moment the user moves away to another page within the parent page, then it will be cleared next time the user comes back to the dashboard.
  function cleanSessionStorageForResource(type: string, name: string) {
    if (type !== ResourceKind.Explore.toString()) return;

    if (resourcesSeen.has(name)) return;
    clearExploreSessionStore(name, EmbedStorageNamespacePrefix);
    resourcesSeen.add(name);
  }

  function handleSelectResource(event: CustomEvent<V1ResourceName>) {
    const newUrl = new URL($page.url);
    newUrl.search = `resource=${event.detail.name}&type=${event.detail.kind}`;
    void goto(newUrl);
  }

  function handleGoHome() {
    const newUrl = new URL($page.url);
    newUrl.search = "";
    void goto(newUrl);
    // To understand why we set `activeResource=null` here, see the comment above
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

{#if navigationEnabled}
  <TopNavigationBarEmbed
    {instanceId}
    {activeResource}
    on:select-resource={handleSelectResource}
    on:go-home={handleGoHome}
  />

  {#if !activeResource}
    <ContentContainer>
      <div class="flex flex-col items-center gap-y-4">
        <DashboardsTable isEmbedded on:select-resource={handleSelectResource} />
      </div>
    </ContentContainer>
  {/if}
{/if}

{#if activeResource}
  {#if activeResource?.kind === ResourceKind.Explore.toString()}
    <ExploreEmbed {instanceId} exploreName={activeResource.name} />
  {:else if activeResource?.kind === ResourceKind.Canvas.toString()}
    <CanvasEmbed
      {instanceId}
      canvasName={activeResource.name}
      {navigationEnabled}
    />
  {:else}
    <UnsupportedKind />
  {/if}
{/if}
