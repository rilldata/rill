<script lang="ts">
  // import { dev } from "$app/environment";
  // import { page } from "$app/stores";
  // import BannerCenter from "@rilldata/web-common/components/banner/BannerCenter.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  // import RepresentingUserBanner from "@rilldata/web-common/features/authentication/RepresentingUserBanner.svelte";
  import ResourceWatcher from "@rilldata/web-common/features/entity-management/ResourceWatcher.svelte";
  // import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  // import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";

  // import ApplicationHeader from "@rilldata/web-common/layout/ApplicationHeader.svelte";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  // import {
  //   initPosthog,
  //   posthogIdentify,
  // } from "@rilldata/web-common/lib/analytics/posthog";
  // import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  // import {
  //   errorEventHandler,
  //   initMetrics,
  // } from "@rilldata/web-common/metrics/initMetrics";
  // import { localServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
  // import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  // import type { Query } from "@tanstack/query-core";
  // import { QueryClientProvider } from "@tanstack/svelte-query";
  // import type { AxiosError } from "axios";
  // import { onMount } from "svelte";
  // import type { LayoutData } from "./$types";
  import "@rilldata/web-common/app.css";

  // import { createAdminServiceGetDeployment } from "@rilldata/web-admin/client";
  // import { onMount } from "svelte";
  // import { page } from "$app/stores";
  // import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";

  export let data;

  // $: ({
  //   params: { organization, project, deployment },
  // } = $page);

  // const deploymentQuery = createAdminServiceGetDeployment();

  // let fetched = false;

  $: ({ instanceId, host, jwt } = data);

  // let removeJavascriptListeners: () => void;

  // $: ({ route } = $page);

  // $: console.log({ instanceId, host, jwt });

  // $: mode = route.id?.includes("(viz)") ? "Preview" : "Developer";
</script>

<RuntimeProvider {instanceId} {host} {jwt} authContext="user">
  <!-- <QueryClientProvider client={queryClient}> -->
  <!-- {#if fetched} -->
  <ResourceWatcher {host} {instanceId}>
    <!-- <div class="body size-full overflow-hidden absolute flex flex-col"> -->
    <!-- {#if data.initialized} -->
    <!-- <BannerCenter /> -->
    <!-- {#if $rillDevCloudFeatures}
          <RepresentingUserBanner />
        {/if} -->
    <!-- <ApplicationHeader {mode} /> -->
    <!-- {/if} -->

    <slot />
  </ResourceWatcher>
  <!-- </div> -->
</RuntimeProvider>

<!-- {/if} -->
<!-- </QueryClientProvider> -->

{#if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
