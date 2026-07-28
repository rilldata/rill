<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import { Button } from "@rilldata/web-common/components/button";
  import { BookmarkIcon, BookmarkPlusIcon } from "lucide-svelte";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceGetProject,
    createAdminServiceListBookmarksInfinite,
    createAdminServiceRemoveBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksInfiniteQueryOptions,
    getAdminServiceListBookmarksQueryKey,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import BookmarkItem from "@rilldata/web-admin/features/project-wide-bookmarks/BookmarkItem.svelte";
  import {
    BookmarksPageSize,
    categorizeBookmarks,
    invalidateBookmarksQuery,
  } from "./utils.ts";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { ExploreSpecProvider } from "@rilldata/web-common/features/dashboards/ExploreSpecProvider.svelte.ts";
  import { page } from "$app/state";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getBookmarkData } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { FilterStateFromUrl } from "@rilldata/web-admin/features/project-wide-bookmarks/FilterStateFromUrl.svelte.ts";
  import type { ParsedBookmark } from "@rilldata/web-admin/features/project-wide-bookmarks/ParsedBookmark.svelte.ts";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let open = $state(false);

  let projectQuery = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let projectId = $derived($projectQuery.data?.project?.id ?? "");

  let bookmarksQuery = $derived(
    createAdminServiceListBookmarksInfinite(
      {
        projectId,
        pageSize: BookmarksPageSize,
      },
      {
        query: {
          enabled: Boolean(projectId),
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    ),
  );
  let allBookmarks = $derived(
    $bookmarksQuery.data?.pages.flatMap((page) => page.bookmarks ?? []) ?? [],
  );
  let categorizedBookmarks = $derived(categorizeBookmarks(allBookmarks));

  const runtimeClient = useRuntimeClient();
  let resource = $derived(getDashboardResourceFromPage(page));

  const filterStateFromUrl = new FilterStateFromUrl(runtimeClient);
  let exploreSpecProvider = $derived(
    resource?.kind === ResourceKind.Explore
      ? new ExploreSpecProvider(runtimeClient, resource.name)
      : undefined,
  );
  let defaultUrlParams = $derived(exploreSpecProvider?.getDefaultUrlParams());

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function onEdit(bookmark: V1Bookmark) {
    console.log("Edit bookmark", bookmark);
    open = false;
  }

  async function onReplace(parsedBookmark: ParsedBookmark) {
    open = false;
    await $bookmarkUpdater.mutateAsync({
      data: {
        bookmarkId: parsedBookmark.bookmark.id,
        displayName: parsedBookmark.bookmark.displayName,
        description: parsedBookmark.bookmark.description,
        shared: parsedBookmark.bookmark.shared,
        urlSearch: getBookmarkData({
          curUrlParams: page.url.searchParams,
          defaultUrlParams,
          filtersOnly: parsedBookmark.filtersOnly,
          absoluteTimeRange: parsedBookmark.absoluteTimeRange,
          selectedTimeRange: filterStateFromUrl?.filterState?.selectedTimeRange,
          selectedComparisonTimeRange:
            filterStateFromUrl?.filterState?.selectedComparisonTimeRange,
        }),
      },
    });
  }

  async function onDelete(bookmark: V1Bookmark) {
    open = false;
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.id,
    });
    eventBus.emit("notification", {
      message: m.bookmark_deleted({
        name: bookmark.displayName ?? "",
      }),
    });
    return invalidateBookmarksQuery(projectId);
  }
</script>

<Dropdown.Root bind:open>
  <Dropdown.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        compact
        square
        type="secondary"
        label="Other bookmark dropdown"
        active={open}
      >
        <BookmarkIcon class="flex-none" size="16px" />
      </Button>
    {/snippet}
  </Dropdown.Trigger>
  <Dropdown.Content class="gap-2 w-[450px]">
    <Dropdown.Group class="grid grid-cols-2 gap-2">
      <Dropdown.Item
        class="justify-center border border-accent-primary-action text-accent-primary-action"
      >
        <BookmarkPlusIcon size="16px" />
        <span>{m.bookmark_current_view()}</span>
      </Dropdown.Item>
      <Dropdown.Item class="justify-center border shadow-xs">
        <HomeBookmarkPlus size="16px" />
        Bookmark as home
      </Dropdown.Item>
    </Dropdown.Group>

    <Dropdown.Separator />

    <!-- TODO: search and sort -->

    <Dropdown.Group>
      <Dropdown.Label class="text-fg-secondary text-[10px] h-6 uppercase">
        {m.bookmark_your_bookmarks()}
      </Dropdown.Label>
      {#each categorizedBookmarks.personal as bookmark (bookmark.id)}
        <BookmarkItem {bookmark} {onEdit} {onReplace} {onDelete} />
      {:else}
        <div class="my-2 text-fg-muted text-center">
          {m.bookmark_no_bookmarks()}
        </div>
      {/each}
    </Dropdown.Group>

    <Dropdown.Separator />

    <Dropdown.Group>
      <Dropdown.Label class="text-fg-secondary">
        <div class="text-[10px] h-4 uppercase">
          {m.bookmark_managed_bookmarks()}
        </div>
        <div class="text-[11px] font-normal">
          {m.bookmark_created_by_admin()}
        </div>
      </Dropdown.Label>
      {#each categorizedBookmarks.shared as bookmark (bookmark.id)}
        <BookmarkItem {bookmark} {onEdit} {onReplace} {onDelete} />
      {:else}
        <div class="my-2 text-fg-muted text-center">
          {m.bookmark_no_shared()}
        </div>
      {/each}
    </Dropdown.Group>
  </Dropdown.Content>
</Dropdown.Root>
