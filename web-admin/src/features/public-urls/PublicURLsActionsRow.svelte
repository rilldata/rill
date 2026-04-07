<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { Trash2Icon, CopyIcon } from "lucide-svelte";
  import DeleteConfirmDialog from "@rilldata/web-common/features/resources/DeleteConfirmDialog.svelte";

  export let id: string;
  export let url: string;
  export let onDelete: (deletedTokenId: string) => void;

  async function handleDelete() {
    onDelete(id);
  }

  function handleCopy() {
    copyToClipboard(url, "Public URL copied to clipboard");
  }

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if url}
      <DropdownMenu.Item
        class="text-fg-primary font-normal flex items-center"
        onclick={handleCopy}
      >
        <CopyIcon size="12px" />
        <span class="ml-2">Copy URL</span>
      </DropdownMenu.Item>
    {/if}
    <DropdownMenu.Item
      class="font-normal flex items-center"
      type="destructive"
      onclick={() => {
        isDeleteConfirmOpen = true;
      }}
    >
      <Trash2Icon size="12px" />
      <span class="ml-2">Delete</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<DeleteConfirmDialog
  bind:open={isDeleteConfirmOpen}
  title="Delete this public URL?"
  description="Recipients of this URL will no longer be able to access it."
  onDelete={handleDelete}
/>
