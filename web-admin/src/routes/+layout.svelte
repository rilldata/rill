<script lang="ts">
  import { page } from "$app/stores";
  import { isAdminServerQuery } from "@rilldata/web-admin/client/utils";
  import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
  import { createUserFacingError } from "@rilldata/web-admin/components/errors/user-facing-errors";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import BillingBannerManager from "@rilldata/web-admin/features/billing/banner/BillingBannerManager.svelte";
  import {
    isBillingUpgradePage,
    isProjectInvitePage,
    isPublicReportPage,
    withinOrganization,
    withinProject,
  } from "@rilldata/web-admin/features/navigation/nav-utils";
  import OrganizationTabs from "@rilldata/web-admin/features/organizations/OrganizationTabs.svelte";
  import { initCloudMetrics } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import BannerCenter from "@rilldata/web-common/components/banner/BannerCenter.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";
  import { isEmbedPage } from "@rilldata/web-common/layout/navigation/navigation-utils.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import { type Query, QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { onMount } from "svelte";
  import ErrorBoundary from "../components/errors/ErrorBoundary.svelte";
  import TopNavigationBar from "../features/navigation/TopNavigationBar.svelte";
  import "@rilldata/web-common/app.css";

  export let data;

  $: ({
    projectPermissions,
    organizationPermissions,
    organizationLogoUrl,
    organizationFaviconUrl,
    planDisplayName,
  } = data);
  $: ({
    params: { organization },
    url: { pathname },
  } = $page);

  // Remember:
  // - https://tkdodo.eu/blog/breaking-react-querys-api-on-purpose#a-bad-api
  // - https://tkdodo.eu/blog/react-query-error-handling#the-global-callbacks
  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => {
    // Add TanStack Query errors to telemetry
    errorEventHandler?.requestErrorEventHandler(error, query);

    // Handle network errors
    // Note: ideally, we'd throw this in the root `+layout.ts` file, but we're blocked by
    // https://github.com/sveltejs/kit/issues/10201
    if (isAdminServerQuery(query) && error.message === "Network Error") {
      errorStore.set(createUserFacingError(null, error.message));
    }
  };

  // The admin server enables some dashboard features like scheduled reports and alerts
  // Set read-only mode so that the user can't edit the dashboard
  featureFlags.set(true, "adminServer", "readOnly");

  let removeJavascriptListeners: () => void;

  initCloudMetrics()
    .then(() => {
      removeJavascriptListeners =
        errorEventHandler?.addJavascriptErrorListeners();
    })
    .catch(console.error);
  initPylonWidget();

  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  $: isEmbed = isEmbedPage($page);

  $: hideTopBar =
    // invite page shouldn't show the top bar because it is considered an onboard step
    isProjectInvitePage($page) ||
    // upgrade callback landing page shouldn't show any rill identifications
    isBillingUpgradePage($page) ||
    // public reports are shared to external users who shouldn't be shown any rill related stuff
    isPublicReportPage($page);
  $: hideBillingManager =
    // billing manager needs organization
    !organization ||
    // invite page shouldn't show the banner since the illusion is that the user is not on cloud yet.
    isProjectInvitePage($page);

  $: withinOnlyOrg = withinOrganization($page) && !withinProject($page);

  function pageContentSizeHandler(node: HTMLElement) {
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        eventBus.emit("page-content-resized", {
          width,
          height,
        });
      }
    });

    resizeObserver.observe(node);

    return {
      destroy() {
        resizeObserver.disconnect();
      },
    };
  }
</script>

<svelte:head>
  <meta content="Rill Cloud" name="description" />
  {#if organizationFaviconUrl}
    <link rel="icon" href={organizationFaviconUrl} />
  {:else}
    <link rel="icon" href="/favicon.png" />
  {/if}
</svelte:head>

<QueryClientProvider client={queryClient}>
  <main
    class="flex flex-col bg-surface"
    class:min-h-screen={!$dynamicHeight}
    class:h-screen={!$dynamicHeight}
    use:pageContentSizeHandler
  >
    <BannerCenter />
    {#if !hideBillingManager}
      <BillingBannerManager {organization} {organizationPermissions} />
    {/if}
    {#if !isEmbed && !hideTopBar}
      <TopNavigationBar
        createMagicAuthTokens={projectPermissions?.createMagicAuthTokens}
        manageProjectMembers={projectPermissions?.manageProjectMembers}
        manageProjectAdmins={projectPermissions?.manageProjectAdmins}
        manageOrgAdmins={organizationPermissions?.manageOrgAdmins}
        manageOrgMembers={organizationPermissions?.manageOrgMembers}
        readProjects={organizationPermissions?.readProjects}
        {organizationLogoUrl}
        {planDisplayName}
      />

      {#if withinOnlyOrg}
        <OrganizationTabs {organization} {organizationPermissions} {pathname} />
      {/if}
    {/if}
    <ErrorBoundary>
      <slot />
    </ErrorBoundary>
  </main>
</QueryClientProvider>

<NotificationCenter />
