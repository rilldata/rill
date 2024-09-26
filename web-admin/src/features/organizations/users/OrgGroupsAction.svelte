<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { Trash2Icon, CopyIcon } from "lucide-svelte";
  import DeleteUserGroupConfirmDialog from "./DeleteUserGroupConfirmDialog.svelte";

  export let id: string;
  export let onDelete: (deletedTokenId: string) => void;

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
    <DropdownMenu.Item
      class="font-normal flex items-center"
      type="destructive"
      on:click={() => {
        isDeleteConfirmOpen = true;
      }}
    >
      <Trash2Icon size="12px" />
      <span class="ml-2">Delete</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<DeleteUserGroupConfirmDialog bind:open={isDeleteConfirmOpen} {id} {onDelete} />
