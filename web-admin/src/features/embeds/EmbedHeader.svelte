<script lang="ts">
  import { isErrorStoreEmpty } from "@rilldata/web-admin/components/errors/error-store";
  import EmbeddedDashboardNavigation from "@rilldata/web-common/features/embeds/EmbeddedDashboardNavigation.svelte";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
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

  export let activeResource: V1ResourceName;
  export let navigationBarEnabled: boolean = true;
  export let navigationEnabled: boolean = navigationBarEnabled;

  const runtimeClient = useRuntimeClient();
  const { twoTieredNavigation, dashboardChat } = featureFlags;

  $: onProjectPage = !activeResource;
  $: showDashboardChat = $dashboardChat && !onProjectPage;

  $: shouldRender = navigationEnabled || showDashboardChat;

  // Dashboard breadcrumb
  $: dashboardsQuery = useValidDashboards(runtimeClient);
  $: ({ data: dashboards } = $dashboardsQuery);
  let currentResource: V1Resource;
  $: currentResource = dashboards?.find(
    (listing) => listing.meta.name.name === activeResource?.name,
  );
  $: currentResourceName = currentResource?.meta?.name?.name;

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

{#if $isErrorStoreEmpty && shouldRender}
  <div class="flex items-center w-full">
    <EmbeddedDashboardNavigation
      {navigationEnabled}
      {navigationBarEnabled}
      {onProjectPage}
      {currentResourceName}
      {breadcrumbOptions}
      twoTiered={$twoTieredNavigation}
    />

    {#if showDashboardChat}
      <div class="flex gap-x-4 items-center">
        <LastRefreshedDate dashboard={activeResource?.name} />
        <ChatToggle open={dashboardChatOpen} actions={dashboardChatActions} />
      </div>
    {/if}
  </div>
{:else}
  <div></div>
{/if}
