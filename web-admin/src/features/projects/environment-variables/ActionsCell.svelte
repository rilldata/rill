<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, Pencil } from "lucide-svelte";
  import EditDialog from "./EditDialog.svelte";
  import DeleteDialog from "./DeleteDialog.svelte";
  import type { VariableNames } from "./types";

  export let id: string;
  export let environment: string;
  export let name: string;
  export let value: string;
  export let variableNames: VariableNames = [];

  let isDropdownOpen = false;
  let isEditDialogOpen = false;
  let isDeleteDialogOpen = false;
</script>

<DropdownMenu.Root bind:open={isDropdownOpen}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="min-w-[95px]">
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

<EditDialog
  bind:open={isEditDialogOpen}
  {id}
  {environment}
  {name}
  {value}
  {variableNames}
/>
<DeleteDialog bind:open={isDeleteDialogOpen} {name} {environment} />
