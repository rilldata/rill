<script lang="ts" module>
  declare module "tanstack-table-8-svelte-5" {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    interface ColumnMeta<TData, TValue> {
      width?: string;
      align?: "left" | "right";
      headerClass?: string;
      cellClass?: string;
    }
  }
</script>

<script lang="ts">
  import { goto } from "$app/navigation";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type {
    ColumnDef,
    SortingState,
    TableOptions,
    Updater,
  } from "tanstack-table-8-svelte-5";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
  } from "tanstack-table-8-svelte-5";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";

  export let data: unknown[] = [];
  // Loosely typed because tanstack-table's ColumnDef discriminated union does
  // not narrow well across component boundaries; callers retain strong typing
  // on their own column definitions.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  export let columns: ColumnDef<any, any>[] = [];
  export let columnVisibility: Record<string, boolean> = {};
  export let kind: string;
  export let toolbar: boolean = true;
  export let initialSorting: SortingState = [];
  /**
   * Whether the caller has applied search/filters to `data` before passing it in.
   * When true and `data` is empty, the "No {kind}s match your search" empty state
   * is shown. Otherwise the regular empty slot is shown.
   */
  export let isFiltered: boolean | undefined = undefined;
  /** If provided, each body row navigates to this href on click. */
  export let getRowHref:
    | ((row: unknown) => string | undefined | null)
    | undefined = undefined;

  let sorting: SortingState = initialSorting;
  function setSorting(updater: Updater<SortingState>) {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }
    options.update((old) => ({
      ...old,
      state: { ...old.state, sorting },
    }));
  }

  const options = writable<TableOptions<unknown>>({
    data: data,
    columns: columns as ColumnDef<unknown, unknown>[],
    enableSorting: true,
    enableFilters: true,
    state: {
      sorting,
      columnVisibility,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);
  setContext("table", table);

  $: options.update((old) => ({ ...old, data }));

  $: rows = $table.getRowModel().rows;
  $: headerGroups = $table.getHeaderGroups();
  $: visibleColumns = $table.getVisibleLeafColumns();
  $: isFilteredEffective =
    isFiltered ?? ($table.getState().globalFilter?.length ?? 0) > 0;

  function navigate(href: string | undefined | null) {
    if (!href) return;
    void goto(href);
  }

  function handleRowClick(e: MouseEvent, href: string | undefined | null) {
    if (!href) return;
    // Don't intercept clicks that originated inside an interactive element
    // (e.g. the actions button). Those handle navigation/behavior themselves.
    const target = e.target as HTMLElement;
    if (target.closest('button, a, [role="menu"], [data-no-row-click]')) return;
    navigate(href);
  }

  function handleRowKeydown(e: KeyboardEvent, href: string | undefined | null) {
    if (!href) return;
    if (e.key === "Enter") {
      e.preventDefault();
      navigate(href);
    }
  }
</script>

<div class="flex flex-col gap-y-3 w-full">
  {#if toolbar}
    <slot name="toolbar" />
  {/if}

  <table class="w-full table-fixed border-collapse">
    <colgroup>
      {#each visibleColumns as column (column.id)}
        <col style:width={column.columnDef.meta?.width} />
      {/each}
    </colgroup>
    <thead>
      {#each headerGroups as group (group.id)}
        <tr class="h-10 border-b border-border">
          {#each group.headers as header (header.id)}
            {@const canSort = header.column.getCanSort()}
            {@const sortDir = header.column.getIsSorted()}
            {@const alignRight =
              header.column.columnDef.meta?.align === "right"}
            <th
              class="px-2 text-fg-muted text-sm font-medium align-middle truncate"
              class:text-left={!alignRight}
              class:text-right={alignRight}
            >
              {#if !header.isPlaceholder}
                {#if canSort}
                  <button
                    type="button"
                    class="inline-flex items-center gap-x-1 max-w-full truncate hover:text-fg-secondary transition-colors"
                    class:text-fg-secondary={!!sortDir}
                    on:click={header.column.getToggleSortingHandler()}
                  >
                    <span class="truncate">
                      <svelte:component
                        this={flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                      />
                    </span>
                    {#if sortDir}
                      <ArrowDown flip={sortDir === "asc"} size="12px" />
                    {/if}
                  </button>
                {:else}
                  <svelte:component
                    this={flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                  />
                {/if}
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </thead>
    <tbody>
      {#if rows.length}
        {#each rows as row (row.id)}
          {@const href = getRowHref?.(row.original)}
          <tr
            class="h-[52px] border-b border-border"
            class:cursor-pointer={!!href}
            class:hover:bg-surface-subtle={!!href}
            role={href ? "link" : undefined}
            tabindex={href ? 0 : undefined}
            on:click={(e) => handleRowClick(e, href)}
            on:keydown={(e) => handleRowKeydown(e, href)}
          >
            {#each row.getVisibleCells() as cell (cell.id)}
              <td
                class="px-2 text-sm text-fg-primary align-middle truncate"
                class:text-right={cell.column.columnDef.meta?.align === "right"}
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
      {:else}
        <tr>
          <td colspan={visibleColumns.length || 1}>
            <div class="text-center py-16">
              {#if isFilteredEffective}
                <div class="flex flex-col gap-y-2 items-center text-sm">
                  <div class="text-fg-secondary font-semibold">
                    No {kind}s match your search
                  </div>
                  <div class="text-fg-secondary">
                    Try adjusting your search terms
                  </div>
                </div>
              {:else}
                <slot name="empty">
                  <div class="text-fg-secondary text-sm font-semibold">
                    You don't have any {kind}s yet
                  </div>
                </slot>
              {/if}
            </div>
          </td>
        </tr>
      {/if}
    </tbody>
  </table>
</div>
