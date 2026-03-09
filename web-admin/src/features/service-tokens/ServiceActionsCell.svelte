<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Pencil, Trash2Icon } from "lucide-svelte";
  import EditServiceDialog from "./EditServiceDialog.svelte";
  import DeleteServiceDialog from "./DeleteServiceDialog.svelte";

  export let name: string;

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
    <DropdownMenu.Content align="end" class="min-w-[95px]">
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
</div>

<EditServiceDialog bind:open={isEditDialogOpen} {name} />
<DeleteServiceDialog bind:open={isDeleteDialogOpen} {name} />
