<script lang="ts">
  import { type BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu";
  import BookmarkFilled from "@rilldata/web-common/components/icons/BookmarkFilled.svelte";
  import BookmarkOutline from "@rilldata/web-common/components/icons/BookmarkOutline.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import FilterFilled from "@rilldata/web-common/components/icons/FilterFilled.svelte";
  import FilterOutline from "@rilldata/web-common/components/icons/FilterOutline.svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { builderActions, getAttrs } from "bits-ui";

  export let bookmark: BookmarkEntry;
  export let readOnly = false;

  export let onClick: () => void = () => {};
  export let onEdit: ((bookmark: BookmarkEntry) => void) | undefined =
    undefined;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;
  // having tooltip for non-home bookmark has an issue where the tooltip persists when moving between edit and delete.
  // since we do not show the edit for home bookmark this is a temporary patch.
  // TODO: figure out why the tooltips persist
  export let showDeleteTooltip = false;

  function editBookmark(e) {
    e.skipSelection = true;
    onEdit?.(bookmark);
  }

  let disableDelete = false;
  async function deleteBookmark(e) {
    disableDelete = true;
    e.skipSelection = true;
    try {
      await onDelete(bookmark);
    } catch {
      // no-op
    }
    disableDelete = false;
  }

  let hovered = false;

  const IconsByType = {
    home: {
      active: HomeBookmark,
      inactive: HomeBookmark,
    },
    filter: {
      active: FilterFilled,
      inactive: FilterOutline,
    },
    complete: {
      active: BookmarkFilled,
      inactive: BookmarkOutline,
    },
  };
  $: icons = bookmark.resource.default
    ? IconsByType.home
    : bookmark.filtersOnly
      ? IconsByType.filter
      : IconsByType.complete;
  $: icon = bookmark.isActive ? icons.active : icons.inactive;
</script>

<DropdownMenuItem class="py-2">
  <div
    class="flex justify-between gap-x-2 w-full"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
    role="menuitem"
    tabindex="-1"
    aria-label={`${bookmark.resource.displayName ?? ""} Bookmark Entry`}
  >
    <a
      href={bookmark.fullUrl}
      class="flex flex-row gap-x-2 w-full min-h-7"
      on:click={onClick}
    >
      <svelte:component this={icon} size="16px" className="text-gray-700" />
      <div class="flex flex-col gap-y-0.5">
        <div
          class="text-xs font-medium text-gray-700 h-4 text-ellipsis overflow-hidden"
        >
          {bookmark.resource.displayName}
        </div>
        {#if bookmark.resource.description}
          <div
            class="text-[11px] font-normal text-gray-500 h-4 text-ellipsis overflow-hidden"
          >
            {bookmark.resource.description}
          </div>
        {/if}
      </div>
    </a>
    {#if !readOnly}
      <div class="flex flex-row justify-end gap-x-2 items-start w-20">
        {#if hovered}
          {#if onEdit}
            <button
              on:click={editBookmark}
              class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
            >
              <EditIcon size="16px" />
            </button>
          {/if}
          <Tooltip.Root portal="body">
            <Tooltip.Trigger asChild let:builder>
              <button
                on:click={deleteBookmark}
                class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
                disabled={disableDelete}
                aria-disabled={disableDelete}
                aria-label="Delete bookmark"
                {...getAttrs([builder])}
                use:builderActions={{ builders: [builder] }}
              >
                <Trash size="16px" />
              </button>
            </Tooltip.Trigger>
            {#if showDeleteTooltip}
              <Tooltip.Content side="bottom">
                Delete {bookmark.resource.default ? "Home " : ""}bookmark
              </Tooltip.Content>
            {/if}
          </Tooltip.Root>
        {/if}
      </div>
    {/if}
  </div>
</DropdownMenuItem>
