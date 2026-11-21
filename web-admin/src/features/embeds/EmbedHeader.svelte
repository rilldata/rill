<script lang="ts">
  import { page } from "$app/stores";
  import { isErrorStoreEmpty } from "@rilldata/web-admin/components/errors/error-store";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import TwoTieredBreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/TwoTieredBreadcrumbItem.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    alertsEnabled,
    featureFlags,
    reportsEnabled,
  } from "@rilldata/web-common/features/feature-flags";
  import LastRefreshedDate from "@rilldata/web-admin/features/dashboards/listing/LastRefreshedDate.svelte";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import {
    dashboardChatActions,
    dashboardChatOpen,
  } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
  import type {
    V1Resource,
    V1ResourceName,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import EmbedReportButton from "@rilldata/web-admin/features/embeds/EmbedReportButton.svelte";
  import EmbedAlertButton from "@rilldata/web-admin/features/embeds/EmbedAlertButton.svelte";
  import EmbedPersonalCanvasButton from "@rilldata/web-admin/features/embeds/EmbedPersonalCanvasButton.svelte";
  import EmbedBookmarks from "@rilldata/web-admin/features/bookmarks/EmbedBookmarks.svelte";
  import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
  import {
    embedBookmarksEnabled,
    getEmbedBookmarkResource,
  } from "@rilldata/web-admin/features/embeds/embed-bookmarks.ts";
  import {
    getEmbedAlertExploreName,
    getEmbedCreateAlertState,
  } from "@rilldata/web-admin/features/embeds/embed-alerts.ts";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { embedPersonalCanvasesEnabled } from "@rilldata/web-admin/features/embeds/embed-personal-canvases.ts";
  import { getPersonalFilteredResourceByName } from "@rilldata/web-admin/features/personal-files/selectors.ts";

  export let activeResource: V1ResourceName;
  export let navigationBarEnabled: boolean = true;

  const runtimeClient = useRuntimeClient();
  const navigationEnabled = !!EmbedStore.getInstance()?.navigationEnabled;
  const { twoTieredNavigation, dashboardChat } = featureFlags;

  $: onProjectPage = !activeResource;
  $: showDashboardChat = $dashboardChat && !onProjectPage;

  // Dashboard breadcrumb
  $: dashboardsQuery = useValidDashboards(runtimeClient);
  $: ({ data: dashboards } = $dashboardsQuery);
  let currentResource: V1Resource;
  $: currentResource = dashboards?.find(
    (listing) => listing.meta.name.name === activeResource?.name,
  );
  $: currentResourceName = currentResource?.meta?.name?.name;

  // Personal canvases aren't valid dashboards for the breadcrumb, so we show their display name instead.
  $: onPersonalCanvasPage = !!getDashboardFromEmbedRoute(
    $page.route.id,
    $page.params,
  )?.personal;
  $: personalCanvasQuery = getPersonalFilteredResourceByName(
    runtimeClient,
    activeResource?.name ?? "",
  );
  $: personalCanvasDisplayName =
    $personalCanvasQuery.data?.canvas?.spec?.displayName ||
    activeResource?.name;

  $: bookmarkResource = getEmbedBookmarkResource($page.route.id, $page.params);

  // Without navigation, only the actions that don't link to other embed pages are shown.
  $: alertExploreName = getEmbedAlertExploreName($page.route.id, $page.params);
  $: createAlertState = getEmbedCreateAlertState(
    runtimeClient,
    alertExploreName,
  );
  $: showAlerts =
    $alertsEnabled && (navigationEnabled || $createAlertState.canCreate);
  $: showReports = $reportsEnabled && navigationEnabled;
  $: showActions =
    $embedPersonalCanvasesEnabled ||
    showDashboardChat ||
    ($embedBookmarksEnabled && !!bookmarkResource) ||
    showAlerts ||
    showReports;

  $: shouldRender = navigationBarEnabled || showActions;

  $: breadcrumbOptions = {
    options: dashboards?.reduce((map, { meta, explore, canvas }) => {
      const name = meta.name.name;
      const isExplore = !!explore;
      return map.set(name.toLowerCase(), {
        label:
          (isExplore
            ? explore?.state?.validSpec?.displayName
            : canvas?.state?.validSpec?.displayName) || name,
        href: `/-/embed/${isExplore ? "explore" : "canvas"}/${name}`,
        preloadData: false,
      });
    }, new Map()),
  };
</script>

{#if shouldRender}
  <div
    class="flex items-center w-full pr-4 py-1 min-h-[2.5rem] bg-surface-subtle"
    class:border-b={!onProjectPage}
  >
    {#if $isErrorStoreEmpty}
      {#if navigationBarEnabled}
        <nav class="flex-1">
          <ol class="flex items-center pl-4">
            {#if !onProjectPage}
              <div class="flex gap-x-2">
                <a
                  class="text-fg-muted hover:text-fg-secondary"
                  href="/-/embed"
                >
                  {m.embed_home()}
                </a>
                <span class="text-fg-muted">/</span>
              </div>
            {/if}

            {#if currentResource}
              {#if $twoTieredNavigation}
                <TwoTieredBreadcrumbItem
                  pathOptions={breadcrumbOptions}
                  current={currentResourceName}
                  isCurrentPage
                />
              {:else}
                <BreadcrumbItem
                  pathOptions={breadcrumbOptions}
                  current={currentResourceName}
                  isCurrentPage
                />
              {/if}
            {:else if onPersonalCanvasPage}
              <span class="text-fg-primary font-medium truncate">
                {personalCanvasDisplayName}
              </span>
            {/if}
          </ol>
        </nav>
      {:else}
        <div class="flex-1"></div>
      {/if}

      <div class="flex gap-x-2 items-center">
        <!-- The last refreshed date is shown alongside the actions, so it is only shown when at least one action is. -->
        {#if showActions && activeResource?.name}
          <LastRefreshedDate dashboard={activeResource.name} />
        {/if}
        {#if $embedPersonalCanvasesEnabled}
          <EmbedPersonalCanvasButton />
        {/if}
        {#if showDashboardChat}
          <ChatToggle open={dashboardChatOpen} actions={dashboardChatActions} />
        {/if}
        {#if $embedBookmarksEnabled && bookmarkResource}
          {#key `${bookmarkResource.kind}:${bookmarkResource.name}`}
            <EmbedBookmarks resource={bookmarkResource} />
          {/key}
        {/if}
        {#if showAlerts}
          <EmbedAlertButton />
        {/if}
        {#if showReports}
          <EmbedReportButton />
        {/if}
      </div>
    {/if}
  </div>
{/if}
