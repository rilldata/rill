<script lang="ts">
  import { page } from "$app/state";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceRemoveBookmark,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils.ts";
  import BookmarkMetadataDialog from "@rilldata/web-admin/features/bookmarks/BookmarkMetadataDialog.svelte";
  import {
    type BookmarkListRow,
    BookmarkTableSortOptions,
    buildBookmarkRows,
    filterBookmarkRows,
  } from "@rilldata/web-admin/features/bookmarks/listing/bookmark-listing-utils.ts";
  import BookmarksTableCompositeCell from "@rilldata/web-admin/features/bookmarks/listing/BookmarksTableCompositeCell.svelte";
  import { RecentlyUsedBookmarks } from "@rilldata/web-admin/features/bookmarks/recently-used-bookmarks.ts";
  import {
    getProjectBookmarksQueryOptions,
    invalidateBookmarkQueries,
  } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { getProjectPermissions } from "@rilldata/web-admin/features/projects/selectors.ts";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import { AlertDialogConfirmation } from "@rilldata/web-common/components/alert-dialog";
  import BookmarkOutline from "@rilldata/web-common/components/icons/BookmarkOutline.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { type RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import { DebouncedRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery } from "@tanstack/svelte-query";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import { writable } from "svelte/store";

  let {
    isPreview = false,
    previewLimit = undefined,
    sortStore: sortStoreProp = undefined,
  }: {
    isPreview?: boolean;
    previewLimit?: number;
    // Lets a parent share the sort state, e.g. the project home page renders the sort dropdown in its heading.
    sortStore?: RuneStore<string>;
  } = $props();

  const defaultSortStore = UrlParamsState.createStringParam(
    "sort",
    BookmarkTableSortOptions[0].value,
  );
  let sortStore = $derived(sortStoreProp ?? defaultSortStore);
  let sortingOption = $derived(
    BookmarkTableSortOptions.find(
      (option) => option.value === sortStore.value,
    ) ?? BookmarkTableSortOptions[0],
  );
  // Secondary sort by name keeps rows with the same primary value in a stable order.
  let sorting = $derived(
    sortingOption.sort.id === "name"
      ? [sortingOption.sort]
      : [sortingOption.sort, { id: "name", desc: false }],
  );

  const searchTextStore = new DebouncedRuneStore(
    UrlParamsState.createStringParam("q"),
    500,
  );

  const runtimeClient = useRuntimeClient();
  let { organization, project } = $derived(page.params);

  const orgAndProjectNameStore = writable({
    organization: page.params.organization,
    project: page.params.project,
  });
  $effect(() => {
    orgAndProjectNameStore.set({ organization, project });
  });

  const user = createAdminServiceGetCurrentUser();
  const bookmarksQuery = createQuery(
    getProjectBookmarksQueryOptions(orgAndProjectNameStore),
  );
  const dashboardsQuery = useDashboards(runtimeClient);
  let permissionsQuery = $derived(getProjectPermissions(organization, project));

  let { data: userData, isPending: isUserPending } = $derived($user);
  let {
    data: bookmarksData,
    isPending: isBookmarksPending,
    isError,
    error,
  } = $derived($bookmarksQuery);
  let { data: dashboardsData } = $derived($dashboardsQuery);
  let { data: permissions } = $derived($permissionsQuery);
  let recentlyUsedBookmarks = $derived(
    new RecentlyUsedBookmarks(organization, project),
  );

  let hasUser = $derived(!!userData?.user);
  // Anonymous viewers have no bookmarks, so only wait on the bookmarks query when there is a user.
  let isLoading = $derived(isUserPending || (hasUser && isBookmarksPending));

  let rows = $derived(
    buildBookmarkRows({
      bookmarks: bookmarksData?.bookmarks ?? [],
      dashboards: dashboardsData ?? [],
      organization,
      project,
      currentUserId: userData?.user?.id,
      manageBookmarks: !!permissions?.manageBookmarks,
      recentlyUsed: recentlyUsedBookmarks.recentlyUsed.value,
    }),
  );
  let filteredRows = $derived(filterBookmarkRows(rows, searchTextStore.value));
  let hasMoreBookmarks = $derived(
    isPreview &&
      previewLimit !== undefined &&
      filteredRows.length > previewLimit,
  );

  let editing = $state<BookmarkListRow | null>(null);
  let deleting = $state<BookmarkListRow | null>(null);

  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function confirmDelete() {
    if (!deleting) return;
    const name = deleting.bookmark.displayName ?? "";
    try {
      await $bookmarkDeleter.mutateAsync({
        bookmarkId: deleting.bookmark.id ?? "",
      });
    } catch (e) {
      eventBus.emit("notification", {
        message: getRpcErrorMessage(e as RpcStatus) ?? String(e),
        type: "error",
      });
      return;
    }
    deleting = null;
    eventBus.emit("notification", {
      message: m.bookmark_deleted({ name }),
    });
    await invalidateBookmarkQueries();
  }

  const columns = [
    {
      id: "composite",
      cell: ({ row }) =>
        renderComponent(BookmarksTableCompositeCell, {
          row: row.original as BookmarkListRow,
          onOpen: (row: BookmarkListRow) =>
            recentlyUsedBookmarks.update(row.bookmark.id),
          onEdit: (row: BookmarkListRow) => (editing = row),
          onDelete: (row: BookmarkListRow) => (deleting = row),
        }),
    },
    {
      id: "name",
      accessorFn: (row: BookmarkListRow) =>
        row.bookmark.displayName?.toLowerCase() ?? "",
    },
    {
      id: "updatedOn",
      accessorFn: (row: BookmarkListRow) =>
        row.bookmark.updatedOn ? new Date(row.bookmark.updatedOn).getTime() : 0,
    },
    {
      id: "dashboard",
      accessorFn: (row: BookmarkListRow) => row.dashboardTitle.toLowerCase(),
    },
    {
      id: "lastUsed",
      accessorFn: (row: BookmarkListRow) => row.lastUsed,
    },
  ];

  const columnVisibility = {
    name: false,
    updatedOn: false,
    dashboard: false,
    lastUsed: false,
  };
</script>

{#if isLoading}
  <div class="m-auto mt-20">
    <DelayedSpinner isLoading={true} size="24px" />
  </div>
{:else if isError}
  <ResourceError kind="bookmark" {error} />
{:else}
  <div class="flex flex-col w-full gap-y-3">
    {#if !isPreview}
      <TableToolbar
        {searchTextStore}
        {sortStore}
        sortOptions={BookmarkTableSortOptions}
      />
    {/if}

    <div class="flex flex-col flex-grow min-w-0">
      <ResourceList
        kind="bookmark"
        data={filteredRows}
        {columns}
        {columnVisibility}
        {sorting}
        toolbar={false}
        maxRows={previewLimit}
        getRowId={(row) => (row as BookmarkListRow).bookmark.id ?? ""}
      >
        <svelte:fragment slot="empty">
          {#if searchTextStore.value}
            <ResourceListEmptyState
              icon={BookmarkOutline}
              message={m.bookmark_list_no_match()}
            />
          {:else}
            <ResourceListEmptyState
              icon={BookmarkOutline}
              message={m.bookmark_list_empty()}
              action={m.bookmark_list_empty_action()}
            />
          {/if}
        </svelte:fragment>
      </ResourceList>

      {#if hasMoreBookmarks}
        <div class="pl-4 py-1">
          <a
            href={`/${organization}/${project}/-/bookmarks`}
            class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
          >
            {m.bookmark_list_see_all()} →
          </a>
        </div>
      {/if}
    </div>
  </div>
{/if}

{#if editing}
  <BookmarkMetadataDialog
    {organization}
    {project}
    bookmark={editing.bookmark}
    href={editing.href}
    onClose={() => (editing = null)}
  />
{/if}

<AlertDialogConfirmation
  open={!!deleting}
  onOpenChange={(open) => {
    if (!open) deleting = null;
  }}
  title={m.bookmark_delete_confirm_title()}
  description={m.bookmark_delete_confirm_description({
    name: deleting?.bookmark.displayName ?? "",
  })}
  confirmLabel={m.common_delete()}
  confirmType="destructive"
  onConfirm={confirmDelete}
/>
