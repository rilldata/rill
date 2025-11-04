<script lang="ts">
  import { page } from "$app/stores";
  import initEmbedPublicAPI from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import { VegaLiteTooltipHandler } from "@rilldata/web-common/components/vega/vega-tooltip.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import ExploreChat from "@rilldata/web-common/features/chat/ExploreChat.svelte";
  import { onMount } from "svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { createIframeRPCHandler } from "@rilldata/web-common/lib/rpc";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  const {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
  } = data;

  const { dashboardChat } = featureFlags;

  // Embedded dashboards communicate directly with the project runtime and do not communicate with the admin server.
  // One by-product of this is that they have no access to control plane features like alerts, bookmarks, and scheduled reports.
  featureFlags.set(false, "adminServer");

  $: activeResource = {
    kind: $page.route.id?.includes("explore")
      ? ResourceKind.Explore
      : ResourceKind.Canvas,
    name: $page.params.name,
  };

  $: showTopBar =
    navigationEnabled ||
    ($dashboardChat &&
      (activeResource?.kind === ResourceKind.Explore.toString() ||
        activeResource?.kind === ResourceKind.MetricsView.toString()));
  $: onProjectPage = !activeResource;

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
    host={runtimeHost}
    {instanceId}
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
        <ExploreChat />
      {/if}
    </div>
  </RuntimeProvider>
{/if}
