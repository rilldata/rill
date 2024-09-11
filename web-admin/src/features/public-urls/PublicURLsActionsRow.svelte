<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

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
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="flex-none">
    <ThreeDot size="16px" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    <DropdownMenu.Item class="text-gray-800 font-normal" on:click={handleCopy}>
      <button on:click={handleCopy}>Copy</button>
    </DropdownMenu.Item>
    <DropdownMenu.Item class="text-gray-800 font-normal">
      <button on:click={handleDelete}>Delete</button>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
