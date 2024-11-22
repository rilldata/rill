<script lang="ts">
  import { writable } from "svelte/store";
  import type { V1MemberUser, V1UserInvite } from "@rilldata/web-admin/client";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";
  import OrgUsersTableRoleCell from "./OrgUsersTableRoleCell.svelte";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import type {
    ColumnDef,
    OnChangeFn,
    SortingState,
    TableOptions,
  } from "@tanstack/svelte-table";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type { InfiniteQueryObserverResult } from "@tanstack/svelte-query";

  interface OrgUser extends V1MemberUser, V1UserInvite {
    invitedBy?: string;
  }

  export let data: OrgUser[];
  export let usersQuery: InfiniteQueryObserverResult;
  export let invitesQuery: InfiniteQueryObserverResult;
  export let currentUserEmail: string;

  const ROW_HEIGHT = 69;
  const OVERSCAN = 5;

  let virtualListEl: HTMLDivElement;
  let sorting: SortingState = [];

  $: safeData = Array.isArray(data) ? data : [];
  $: {
    if (safeData) {
      options.update((old) => ({
        ...old,
        data: safeData,
      }));
    }
  }

  const columns: ColumnDef<OrgUser, any>[] = [
    {
      accessorKey: "user",
      header: "User",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableUserCompositeCell, {
          name: row.original.userName ?? row.original.email,
          email: row.original.userEmail,
          pendingAcceptance: Boolean(row.original.invitedBy),
          isCurrentUser: row.original.userEmail === currentUserEmail,
          photoUrl: row.original.userPhotoUrl,
        }),
      meta: {
        widthPercent: 5,
      },
    },
    {
      accessorKey: "roleName",
      header: "Role",
      cell: ({ row }) =>
        flexRender(OrgUsersTableRoleCell, {
          email: row.original.userEmail,
          role: row.original.roleName,
          isCurrentUser: row.original.userEmail === currentUserEmail,
        }),
      meta: {
        widthPercent: 5,
        marginLeft: "8px",
      },
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableActionsCell, {
          email: row.original.userEmail,
          isCurrentUser: row.original.userEmail === currentUserEmail,
        }),
      meta: {
        widthPercent: 0,
      },
    },
  ];

  const setSorting: OnChangeFn<SortingState> = (updater) => {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }

    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        sorting,
      },
    }));
  };

  const options = writable<TableOptions<OrgUser>>({
    data: safeData,
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);

  $: rows = $table.getRowModel().rows;

  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });

  $: {
    const hasNextPage = usersQuery.hasNextPage || invitesQuery.hasNextPage;

    $virtualizer.setOptions({
      count: hasNextPage ? safeData.length + 1 : safeData.length,
    });

    const [lastItem] = [...$virtualizer.getVirtualItems()].reverse();

    if (
      lastItem &&
      lastItem.index > safeData.length - 1 &&
      hasNextPage &&
      !usersQuery.isFetchingNextPage &&
      !invitesQuery.isFetchingNextPage
    ) {
      if (usersQuery.hasNextPage) {
        usersQuery.fetchNextPage();
      }
      if (invitesQuery.hasNextPage) {
        invitesQuery.fetchNextPage();
      }
    }
  }

  $: dynamicTableMaxHeight = data.length > 12 ? `calc(100dvh - 300px)` : "auto";
</script>

<!-- FIXME: hoist this to a InfiniteScrollTable component -->
<div
  class={`list scroll-container ${dynamicTableMaxHeight}`}
  bind:this={virtualListEl}
  style:max-height={dynamicTableMaxHeight}
>
  <div class="table-wrapper" style="position: relative;">
    <table>
      <thead>
        {#each $table.getHeaderGroups() as headerGroup}
          <tr class="h-10">
            {#each headerGroup.headers as header (header.id)}
              {@const widthPercent = header.column.columnDef.meta?.widthPercent}
              {@const marginLeft = header.column.columnDef.meta?.marginLeft}
              <th
                colSpan={header.colSpan}
                style={`width: ${widthPercent}%;`}
                class="px-4 py-2 text-left"
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  <div
                    style={`margin-left: ${marginLeft};`}
                    class:cursor-pointer={header.column.getCanSort()}
                    class:select-none={header.column.getCanSort()}
                    class="font-semibold text-gray-500 flex flex-row items-center gap-x-1"
                  >
                    <svelte:component
                      this={flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                    />
                    {#if header.column.getIsSorted().toString() === "asc"}
                      <span>
                        <ArrowDown flip size="12px" />
                      </span>
                    {:else if header.column.getIsSorted().toString() === "desc"}
                      <span>
                        <ArrowDown size="12px" />
                      </span>
                    {/if}
                  </div>
                {/if}
              </th>
            {/each}
          </tr>
        {/each}
      </thead>
      <tbody>
        {#if $table.getRowModel().rows.length === 0}
          <tr>
            <td
              colspan={columns.length}
              class="px-4 py-4 text-center text-gray-500"
            >
              No users found
            </td>
          </tr>
        {:else}
          {#each $virtualizer.getVirtualItems() as virtualRow, idx (virtualRow.index)}
            <tr
              style="height: {virtualRow.size}px; transform: translateY({virtualRow.start -
                idx * virtualRow.size}px);"
            >
              {#each rows[virtualRow.index]?.getVisibleCells() ?? [] as cell (cell.id)}
                <td
                  class={`px-4 py-2 max-w-[200px] truncate ${cell.column.id === "actions" ? "w-1" : ""}`}
                  data-label={cell.column.columnDef.header}
                >
                  <svelte:component
                    this={flexRender(
                      cell.column.columnDef.cell,
                      cell.getContext(),
                    )}
                  />
                </td>
              {/each}
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</div>

<style lang="postcss">
  table {
    @apply border-separate border-spacing-0 w-full;
  }
  table th,
  table td {
    @apply border-b border-gray-200;
  }
  thead {
    @apply sticky top-0 z-30 bg-white;
  }
  thead tr th {
    @apply border-t border-gray-200;
  }
  thead tr th:first-child {
    @apply border-l;
    @apply rounded-tl-sm;
  }
  thead tr th:last-child {
    @apply border-r;
    @apply rounded-tr-sm;
  }
  thead tr:last-child th {
    @apply border-b;
  }
  tbody tr:first-child {
    @apply border-t-0;
  }
  tbody td:first-child {
    @apply border-l;
  }
  tbody td:last-child {
    @apply border-r;
  }
  tbody tr:last-child td:first-child {
    @apply rounded-bl-sm;
  }
  tbody tr:last-child td:last-child {
    @apply rounded-br-sm;
  }
  .scroll-container {
    width: 100%;
    overflow-y: auto;
  }
</style>
