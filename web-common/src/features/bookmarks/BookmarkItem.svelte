<script lang="ts">
  import {
    createAdminServiceRemoveBookmark,
    type V1Bookmark,
  } from "@rilldata/web-admin/client";
  import { PencilIcon, TrashIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  export let bookmark: V1Bookmark;

  const dispatch = createEventDispatcher();
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function deleteBookmark() {
    await $bookmarkDeleter.mutateAsync({ bookmarkId: bookmark.id as string });
  }
</script>

<div class="flex flex-row gap-2">
  <div
    class="flex flex-col gap-2"
    on:click={() => dispatch("select", bookmark)}
  >
    <div class="text-sm text-gray-700">{bookmark.displayName}</div>
    {#if bookmark.description}
      <div class="text-sm text-gray-500 h-8 text-ellipsis">
        {bookmark.description}
      </div>
    {/if}
  </div>
  <button on:click={() => dispatch("edit", bookmark)}>
    <PencilIcon size="16px" />
  </button>
  <button on:click={deleteBookmark}><TrashIcon size="16px" /></button>
</div>
