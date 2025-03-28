<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getHomeBookmarkURLParams } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;
  export let metricsViewName: string;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);

  $: homeBookmarkURLParams = getHomeBookmarkURLParams(
    organization,
    project,
    instanceId,
    metricsViewName,
    exploreName,
  );
  $: href =
    $page.url.pathname +
    ($homeBookmarkURLParams ? "?" + $homeBookmarkURLParams : "");

  function goToDashboardHome() {
    clearExploreSessionStore(exploreName, `${organization}__${project}__`);
    return goto(href);
  }
</script>

<Tooltip.Root portal="body">
  <Tooltip.Trigger>
    <Button compact type="secondary" on:click={goToDashboardHome}>
      <HomeBookmark size="16px" />
    </Button>
  </Tooltip.Trigger>
  <Tooltip.Content side="bottom">Return to dashboard home</Tooltip.Content>
</Tooltip.Root>
