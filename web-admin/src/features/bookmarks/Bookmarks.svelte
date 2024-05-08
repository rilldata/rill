<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { createBookmarkApplier } from "@rilldata/web-admin/features/bookmarks/applyBookmark";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuContent.svelte";
  import CreateBookmarkDialog from "@rilldata/web-admin/features/bookmarks/CreateBookmarkDialog.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import EditBookmarkDialog from "@rilldata/web-admin/features/bookmarks/EditBookmarkDialog.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon } from "lucide-svelte";

  $: metricsViewName = $page.params.dashboard;

  let createBookmark = false;
  let editBookmark = false;
  let bookmark: BookmarkEntry;

  $: bookmarkApplier = createBookmarkApplier(
    $runtime?.instanceId,
    metricsViewName,
  );

  $: dashboardStore = useDashboardStore(metricsViewName);
  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const queryClient = useQueryClient();
  $: homeBookmarkModifier = createHomeBookmarkModifier($runtime?.instanceId);
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function selectBookmark(bookmark: BookmarkEntry) {
    bookmarkApplier(bookmark.resource);
  }

  async function createHomeBookmark() {
    await homeBookmarkModifier(
      getBookmarkDataForDashboard($dashboardStore, false, false),
    );
    notifications.send({
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
      }),
    );
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation?
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id,
    });
    notifications.send({
      message: `Bookmark ${bookmark.resource.displayName} deleted`,
    });
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
      }),
    );
  }

  let open = false;
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} compact type="secondary">
      <BookmarkIcon
        class="inline-flex"
        fill={open ? "black" : "none"}
        size="16px"
      />
    </Button>
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => (createBookmark = true)}
    on:create-home={() => createHomeBookmark()}
    on:delete={({ detail }) => deleteBookmark(detail)}
    on:edit={({ detail }) => {
      editBookmark = true;
      bookmark = detail;
    }}
    on:select={({ detail }) => selectBookmark(detail)}
  />
</DropdownMenu>

<CreateBookmarkDialog bind:open={createBookmark} {metricsViewName} />

{#if bookmark}
  {#key bookmark.resource.id}
    <EditBookmarkDialog {bookmark} bind:open={editBookmark} {metricsViewName} />
  {/key}
{/if}
