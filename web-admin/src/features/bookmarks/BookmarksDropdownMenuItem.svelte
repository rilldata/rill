<script lang="ts">
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { BookmarkIcon } from "lucide-svelte";

  export let bookmark: BookmarkEntry;
  export let readOnly = false;

  export let onEdit: (bookmark: BookmarkEntry) => void;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;

  function editBookmark(e) {
    e.skipSelection = true;
    onEdit(bookmark);
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
    <a href={bookmark.url} class="flex flex-row gap-x-2 w-full min-h-7">
      {#if bookmark.resource.default}
        <HomeBookmark size="16px" />
      {:else if bookmark.filtersOnly}
        <Filter size="16px" />
      {:else}
        <BookmarkIcon size="16px" aria-label="Bookmark Icon" />
      {/if}
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
          <button
            on:click={editBookmark}
            class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
          >
            <EditIcon size="16px" />
          </button>
          <button
            on:click={deleteBookmark}
            class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
            disabled={disableDelete}
            aria-disabled={disableDelete}
          >
            <Trash size="16px" />
          </button>
        {/if}
      </div>
    {/if}
  </div>
</DropdownMenuItem>
