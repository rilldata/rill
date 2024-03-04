<script lang="ts">
  import { createAdminServiceRemoveBookmark } from "@rilldata/web-admin/client";
  import type { BookmarkEntry } from "@rilldata/web-common/features/bookmarks/selectors";
  import { PencilIcon, TrashIcon, BookmarkIcon } from "lucide-svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu/index";

  export let bookmark: BookmarkEntry;

  const dispatch = createEventDispatcher();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function selectBookmark(e) {
    if (e.skipSelection) return;
    dispatch("select", bookmark);
  }

  function editBookmark(e) {
    e.skipSelection = true;
    dispatch("edit", bookmark);
  }

  async function deleteBookmark(e) {
    e.skipSelection = true;
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id as string,
    });
  }

  let hovered = false;
</script>

<DropdownMenuItem>
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
      <div class="flex flex-col">
        <div
          class="text-xs font-medium text-gray-700 h-5 text-ellipsis overflow-hidden"
        >
          {bookmark.resource.displayName}
        </div>
        {#if bookmark.resource.description}
          <div
            class="text-[11px] font-normal text-gray-500 h-5 text-ellipsis overflow-hidden"
          >
            {bookmark.resource.description}
          </div>
        {/if}
      </div>
    </div>
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
  </div>
</DropdownMenuItem>
