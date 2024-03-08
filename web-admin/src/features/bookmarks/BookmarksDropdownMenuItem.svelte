<script lang="ts">
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { PencilIcon, TrashIcon, BookmarkIcon } from "lucide-svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu";

  export let bookmark: BookmarkEntry;
  export let readOnly = false;

  const dispatch = createEventDispatcher();

  function selectBookmark(e) {
    if (e.skipSelection) return;
    dispatch("select", bookmark);
  }

  function editBookmark(e) {
    e.skipSelection = true;
    dispatch("edit", bookmark);
  }

  function deleteBookmark(e) {
    e.skipSelection = true;
    dispatch("delete", bookmark);
  }

  let hovered = false;
</script>

<DropdownMenuItem class="py-1">
  <div
    class="flex justify-between gap-x-2 w-full"
    on:click={selectBookmark}
    on:keydown={(e) => e.key === "Enter" && e.currentTarget.click()}
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
    role="menuitem"
    tabindex="0"
  >
    <div class="flex flex-row gap-x-2">
      {#if bookmark.resource.default}
        <HomeBookmark size="16px" />
      {:else if bookmark.filtersOnly}
        <Filter size="16px" />
      {:else}
        <BookmarkIcon size="16px" />
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
    </div>
    {#if !readOnly}
      <div class="flex flex-row justify-end gap-x-2 items-start w-20">
        {#if hovered}
          <button on:click={editBookmark}>
            <PencilIcon size="16px" />
          </button>
          <button on:click={deleteBookmark}>
            <TrashIcon size="16px" />
          </button>
        {/if}
      </div>
    {/if}
  </div>
</DropdownMenuItem>
