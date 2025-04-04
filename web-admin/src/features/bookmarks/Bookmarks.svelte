<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListBookmarks,
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import BookmarkDialog from "@rilldata/web-admin/features/bookmarks/BookmarkDialog.svelte";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuContent.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon } from "lucide-svelte";

  export let metricsViewName: string;
  export let exploreName: string;

  let showDialog = false;
  let bookmark: BookmarkEntry | null = null;

  const { validSpecStore } = getStateManagers();

  $: exploreState = useExploreState(exploreName);
  $: projectId = useProjectId($page.params.organization, $page.params.project);
  const userResp = createAdminServiceGetCurrentUser();
  $: bookamrksResp = createAdminServiceListBookmarks(
    {
      projectId: $projectId.data,
      resourceKind: ResourceKind.Explore,
      resourceName: exploreName,
    },
    {
      query: {
        enabled: !!$projectId.data && !!$userResp.data.user,
      },
    },
  );

  const queryClient = useQueryClient();
  $: homeBookmarkModifier = createHomeBookmarkModifier(exploreName);
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function createHomeBookmark() {
    await homeBookmarkModifier(
      getBookmarkDataForDashboard(
        $exploreState,
        $validSpecStore.data?.explore ?? {},
      ),
      $projectId.data,
      $bookamrksResp.data?.bookmarks ?? [],
    );
    eventBus.emit("notification", {
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    });
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation?
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id,
    });
    eventBus.emit("notification", {
      message: `Bookmark ${bookmark.resource.displayName} deleted`,
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    });
  }

  let open = false;
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button
      builders={[builder]}
      compact
      type="secondary"
      label="Bookmark dropdown"
    >
      <BookmarkIcon
        class="inline-flex"
        fill={open ? "black" : "none"}
        size="16px"
      />
    </Button>
  </DropdownMenuTrigger>
  <BookmarksContent
    onCreate={(isHome) => {
      isHome ? createHomeBookmark() : (showDialog = true);
    }}
    onEdit={(editingBookmark) => {
      showDialog = true;
      bookmark = editingBookmark;
    }}
    onDelete={deleteBookmark}
    {metricsViewName}
    {exploreName}
  />
</DropdownMenu>

{#if showDialog}
  <BookmarkDialog
    {bookmark}
    {metricsViewName}
    {exploreName}
    onClose={() => {
      showDialog = false;
      bookmark = null;
    }}
  />
{/if}
