<script lang="ts">
  import { getHomeBookmarkButtonUrl } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let projectId: string;
  export let project: string;
  export let metricsViewName: string;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);

  $: homeBookmarkUrl = getHomeBookmarkButtonUrl(
    projectId,
    instanceId,
    metricsViewName,
    exploreName,
  );

  function goToDashboardHome() {
    // Without clearing sessions empty url will load that instead
    clearExploreSessionStore(exploreName, `${organization}__${project}__`);
  }
</script>

<Tooltip.Root portal="body">
  <Tooltip.Trigger>
    <Button
      type="link"
      compact
      preload={false}
      href={$homeBookmarkUrl}
      on:click={goToDashboardHome}
      class="border border-primary-300"
    >
      <HomeBookmark size="16px" />
    </Button>
  </Tooltip.Trigger>
  <Tooltip.Content side="bottom">Return to dashboard home</Tooltip.Content>
</Tooltip.Root>
