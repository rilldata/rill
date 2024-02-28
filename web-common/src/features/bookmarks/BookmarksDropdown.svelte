<script lang="ts">
    import {page} from "$app/stores";
    import {createAdminServiceCreateBookmark, type V1Bookmark} from "@rilldata/web-admin/client";
    import {DropdownMenu, DropdownMenuTrigger,} from "@rilldata/web-common/components/dropdown-menu";
    import {createApplyBookmark} from "@rilldata/web-common/features/bookmarks/applyBookmark";
    import BookmarksContent from "@rilldata/web-common/features/bookmarks/BookmarksContent.svelte";
    import CreateBookmarkDialog from "@rilldata/web-common/features/bookmarks/CreateBookmarkDialog.svelte";
    import EditBookmarkDialog from "@rilldata/web-common/features/bookmarks/EditBookmarkDialog.svelte";
    import {getBookmarkForDashboard} from "@rilldata/web-common/features/bookmarks/getBookmarkForDashboard";
    import {useProjectId} from "@rilldata/web-common/features/bookmarks/selectors";
    import {useDashboardStore} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
    import {runtime} from "@rilldata/web-common/runtime-client/runtime-store";
    import {BookmarkIcon} from "lucide-svelte";

    $: metricsViewName = $page.params.dashboard;

  let createBookmark = false;
  let editBookmark = false;
  let bookmark: V1Bookmark;

  $: bookmarkApplier = createApplyBookmark(
    $runtime?.instanceId,
    metricsViewName,
  );

  $: dashboardStore = useDashboardStore(metricsViewName);
  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const bookmarkCreator = createAdminServiceCreateBookmark();

  function onSelect(bookmark: V1Bookmark) {
    bookmarkApplier(bookmark);
  }

  async function createHomeBookmark() {
    await $bookmarkCreator.mutateAsync({
      data: {
        displayName: "Home",
        description: "Default bookmark",
        projectId: $projectId.data ?? "",
        dashboardName: metricsViewName,
        shared: true,
        default: true,
        data: getBookmarkForDashboard(
          $dashboardStore,
          false,
          false,
        ),
      },
    });
  }
</script>

<DropdownMenu>
  <DropdownMenuTrigger>
    <BookmarkIcon class="inline-flex"/>
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => (createBookmark = true)}
    on:create-home={() => createHomeBookmark()}
    on:edit={({ detail }) => {
      editBookmark = true;
      bookmark = detail;
    }}
    on:select={({ detail }) => onSelect(detail)}
  />
</DropdownMenu>

<CreateBookmarkDialog bind:open={createBookmark} {metricsViewName}/>

{#if bookmark}
  <EditBookmarkDialog {bookmark} bind:open={editBookmark} {metricsViewName}/>
{/if}
