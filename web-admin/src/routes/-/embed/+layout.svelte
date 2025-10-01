<script lang="ts">
  import { page } from "$app/stores";
  import initEmbedPublicAPI from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
  import TopNavigationBarEmbed from "@rilldata/web-admin/features/embeds/TopNavigationBarEmbed.svelte";
  import { VegaLiteTooltipHandler } from "@rilldata/web-common/components/vega/vega-tooltip.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
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

  $: activeResource = {
    kind: $page.route.id?.includes("explore")
      ? ResourceKind.Explore
      : ResourceKind.Canvas,
    name: $page.params.name,
  };

  onMount(() => {
    createIframeRPCHandler();
    VegaLiteTooltipHandler.resetElement();

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
    {#if navigationEnabled}
      <TopNavigationBarEmbed {instanceId} {activeResource} />
    {/if}

    <slot />
  </RuntimeProvider>
{/if}
