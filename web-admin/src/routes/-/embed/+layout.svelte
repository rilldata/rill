<script lang="ts">
  import { beforeNavigate, onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    getDashboardFromEmbedRoute,
    isDifferentDashboard,
  } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
  import initEmbedPublicAPI from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { VegaLiteTooltipHandler } from "@rilldata/web-common/components/vega/vega-tooltip.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
  import {
    createIframeRPCHandler,
    emitNotification,
  } from "@rilldata/web-common/lib/rpc";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  const {
    instanceId,
    missingRequireParams,
    navigationEnabled,
    runtimeHost,
    accessToken,
  } = data;

  const { dashboardChat } = featureFlags;

  // Embedded dashboards communicate directly with the project runtime and do not communicate with the admin server.
  // One by-product of this is that they have no access to control plane features like alerts, bookmarks, and scheduled reports.
  featureFlags.set(false, "adminServer");

  // Extract active resource info from current route
  // Falls back to Canvas if route doesn't match a dashboard pattern (e.g., project page)
  $: activeResource = getDashboardFromEmbedRoute(
    $page.route.id,
    $page.params,
  ) ?? {
    kind: ResourceKind.Canvas,
    name: $page.params.name,
  };

  $: showTopBar =
    navigationEnabled ||
    ($dashboardChat &&
      (activeResource?.kind === ResourceKind.Explore.toString() ||
        activeResource?.kind === ResourceKind.MetricsView.toString()));
  $: onProjectPage = !activeResource;

  // Suppress browser back/forward
  beforeNavigate((nav) => {
    if (!navigationEnabled) {
      if (nav.type === "popstate") {
        nav.cancel();
      }
    }
  });

  onNavigate(({ from, to }) => {
    if (!navigationEnabled) return;

    const fromDashboard = from
      ? getDashboardFromEmbedRoute(from.route.id, from.params)
      : null;
    const toDashboard = to
      ? getDashboardFromEmbedRoute(to.route.id, to.params)
      : null;

    if (
      fromDashboard &&
      toDashboard &&
      isDifferentDashboard(fromDashboard, toDashboard)
    ) {
      emitNotification("navigation", {
        from: fromDashboard.name,
        to: toDashboard.name,
      });
    } else if (!fromDashboard && toDashboard) {
      // Navigating from listing page to a dashboard
      emitNotification("navigation", {
        from: "dashboardListing",
        to: toDashboard.name,
      });
    } else if (fromDashboard && !toDashboard) {
      // Navigating from a dashboard to the listing page
      emitNotification("navigation", {
        from: fromDashboard.name,
        to: "dashboardListing",
      });
    }
  });

  onMount(() => {
    createIframeRPCHandler();
    void waitUntil(() => VegaLiteTooltipHandler.resetElement(), 5000, 100);

    return initEmbedPublicAPI();
  });
</script>

<svelte:head>
  {#if activeResource}
    <title>{activeResource.name} - Rill</title>
  {:else}
    <title>Rill</title>
  {/if}
</svelte:head>

{#if missingRequireParams.length}
  <ErrorPage
    header={`Missing required param(s) ${missingRequireParams
      .map((p) => '"' + p + '"')
      .join(",")}`}
    fatal
  />
{:else}
  <RuntimeProvider
    {instanceId}
    host={runtimeHost}
    jwt={accessToken}
    authContext="embed"
  >
    {#if showTopBar}
      <div
        class="flex items-center w-full pr-4 py-1 min-h-[2.5rem]"
        class:border-b={!onProjectPage}
      >
        <TopNavigationBarEmbed
          {instanceId}
          {activeResource}
          {navigationEnabled}
        />
      </div>
    {/if}

    <div class="flex h-full overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <slot />
      </div>
      {#if $dashboardChat && activeResource?.kind === ResourceKind.Explore}
        <DashboardChat />
      {/if}
    </div>
  </RuntimeProvider>
{/if}
