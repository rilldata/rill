<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, Pencil } from "lucide-svelte";
  import EditDialog from "./EditDialog.svelte";
  import DeleteDialog from "./DeleteDialog.svelte";

  export let name: string;
  export let environment: string;

  let isDropdownOpen = false;
  let isDeleteDialogOpen = false;
  let isEditDialogOpen = false;
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
      on:click={() => {
        isEditDialogOpen = true;
      }}
    >
      <Pencil size="12px" />
      <span class="ml-2">Edit</span>
    </DropdownMenu.Item>
    <DropdownMenu.Item
      class="font-normal flex items-center"
      type="destructive"
      on:click={() => {
        isDeleteDialogOpen = true;
      }}
    >
      <Trash2Icon size="12px" />
      <span class="ml-2">Delete</span>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<EditDialog bind:open={isEditDialogOpen} />
<DeleteDialog bind:open={isDeleteDialogOpen} {name} {environment} />
