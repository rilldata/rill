<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { createBookmarkApplier } from "@rilldata/web-admin/features/bookmarks/applyBookmark";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksContent.svelte";
  import CreateBookmarkDialog from "@rilldata/web-admin/features/bookmarks/CreateBookmarkDialog.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import EditBookmarkDialog from "@rilldata/web-admin/features/bookmarks/EditBookmarkDialog.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import {
    type BookmarkEntry,
    useProjectId,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
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
  const homeBookmarkModifier = createHomeBookmarkModifier();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function selectBookmark(bookmark: BookmarkEntry) {
    bookmarkApplier(bookmark.resource);
  }

  async function createHomeBookmark() {
    await homeBookmarkModifier(
      getBookmarkDataForDashboard($dashboardStore, false, false),
    );
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
      }),
    );
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id as string,
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

<DropdownMenu bind:open>
  <DropdownMenuTrigger>
    <BookmarkIcon class="inline-flex" fill={open ? "black" : "none"} />
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
