<script lang="ts">
  import { onDestroy } from "svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";

  export let id: string;
  export let url: string;
  export let onDelete: (deletedTokenId: string) => void;

  let copied = false;
  let copyTimer: ReturnType<typeof setTimeout>;

  const COPIED_TIMER = 1_500;

  function handleCopy() {
    navigator.clipboard.writeText(url);
    copied = true;

    if (copyTimer) clearTimeout(copyTimer);

    copyTimer = setTimeout(() => {
      copied = false;
    }, COPIED_TIMER);
  }

  async function handleDelete() {
    try {
      onDelete(id);
    } catch (error) {
      console.error("Failed to delete magic auth token:", error);
    }
  }

  onDestroy(() => {
    if (copyTimer) clearTimeout(copyTimer);
  });
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="flex-none">
    <ThreeDot size="16px" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    <DropdownMenu.Item class="text-gray-800 font-normal" on:click={handleCopy}>
      {#if copied}
        <button>Copied</button>
      {:else if url}
        <button on:click={handleCopy}> Copy </button>
      {/if}
    </DropdownMenu.Item>
    <DropdownMenu.Item class="text-gray-800 font-normal">
      <button on:click={handleDelete}>Delete</button>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
