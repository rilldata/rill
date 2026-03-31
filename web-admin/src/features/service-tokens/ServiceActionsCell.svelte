<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { KeyRound, Pencil, Trash2Icon } from "lucide-svelte";
  import EditServiceDialog from "./EditServiceDialog.svelte";
  import DeleteServiceDialog from "./DeleteServiceDialog.svelte";

  export let name: string;
  export let onManageTokens: (name: string) => void;

  let isDropdownOpen = false;
  let isEditDialogOpen = false;
  let isDeleteDialogOpen = false;
</script>

<div class="flex items-center">
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="18px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="min-w-[140px]">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        onclick={() => onManageTokens(name)}
      >
        <KeyRound size="12px" />
        <span class="ml-2">Manage tokens</span>
      </DropdownMenu.Item>
      <DropdownMenu.Item
        class="font-normal flex items-center"
        onclick={() => {
          isEditDialogOpen = true;
        }}
      >
        <Pencil size="12px" />
        <span class="ml-2">Edit</span>
      </DropdownMenu.Item>
      <DropdownMenu.Item
        class="font-normal flex items-center"
        type="destructive"
        onclick={() => {
          isDeleteDialogOpen = true;
        }}
      >
        <Trash2Icon size="12px" />
        <span class="ml-2">Delete</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

<EditServiceDialog bind:open={isEditDialogOpen} {name} />
<DeleteServiceDialog bind:open={isDeleteDialogOpen} {name} />
