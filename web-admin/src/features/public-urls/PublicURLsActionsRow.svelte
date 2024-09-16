<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { Trash2Icon, CopyIcon } from "lucide-svelte";

  export let id: string;
  export let url: string;
  export let onDelete: (deletedTokenId: string) => void;

  function handleCopy() {
    copyToClipboard(url, "Public URL copied to clipboard");
  }

  async function handleDelete() {
    try {
      onDelete(id);
    } catch (error) {
      console.error("Failed to delete magic auth token:", error);
    }
  }

  let isOpen = false;
</script>

<DropdownMenu.Root bind:open={isOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if url}
      <DropdownMenu.Item
        class="text-gray-800 font-normal flex items-center"
        on:click={handleCopy}
      >
        <CopyIcon size="12px" />
        <span class="ml-2">Copy URL</span>
      </DropdownMenu.Item>
    {/if}
    <DropdownMenu.Item
      class="font-normal flex items-center"
      type="destructive"
      on:click={handleDelete}
    >
      <Trash2Icon size="12px" />
      <span class="ml-2">Delete</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
