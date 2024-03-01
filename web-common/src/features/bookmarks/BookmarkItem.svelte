<script lang="ts">
  import {
    createAdminServiceRemoveBookmark,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import { PencilIcon, TrashIcon, BookmarkIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu/index";

  export let bookmark: V1Bookmark;

  const dispatch = createEventDispatcher();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function deleteBookmark() {
    await $bookmarkDeleter.mutateAsync({ bookmarkId: bookmark.id as string });
  }

  let hovered = false;
</script>

<DropdownMenuItem>
  <div
    class="flex justify-between gap-x-2 w-full"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
    role="menuitem"
    tabindex="0"
  >
    <div
      class="flex flex-row gap-x-2"
      on:click={() => dispatch("select", bookmark)}
      on:keydown={(e) => e.key === "Enter" && e.currentTarget.click()}
      role="button"
      tabindex="0"
    >
      <div class="pt-0.5">
        {#if bookmark.default}
          <HomeBookmark size="16px" />
        {:else}
          <BookmarkIcon size="16px" />
        {/if}
      </div>
      <div class="flex flex-col">
        <div class="text-sm text-gray-700 h-5 text-ellipsis overflow-hidden">
          {bookmark.displayName}
        </div>
        {#if bookmark.description}
          <div class="text-sm text-gray-500 h-5 text-ellipsis overflow-hidden">
            {bookmark.description}
          </div>
        {/if}
      </div>
    </div>
    <div class="flex flex-row justify-end gap-x-2 items-start pt-1 w-20">
      {#if hovered}
        <button on:click={() => dispatch("edit", bookmark)}>
          <PencilIcon size="16px" />
        </button>
        <button on:click={deleteBookmark}><TrashIcon size="16px" /></button>
      {/if}
    </div>
  </div>
</DropdownMenuItem>
