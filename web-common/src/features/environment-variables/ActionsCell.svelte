<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon, Pencil } from "lucide-svelte";
  import EditEnvDialog from "./EditEnvDialog.svelte";
  import DeleteEnvDialog from "./DeleteEnvDialog.svelte";
  import type { EnvVariable } from "./types";

  export let keyName: string;
  export let value: string;
  export let existingVariables: EnvVariable[] = [];
  export let onSave: (oldKey: string, key: string, value: string) => void;
  export let onDelete: (key: string) => void;

  let isDropdownOpen = false;
  let isEditDialogOpen = false;
  let isDeleteDialogOpen = false;

  function handleSave(event: CustomEvent<{ oldKey: string; key: string; value: string }>) {
    onSave(event.detail.oldKey, event.detail.key, event.detail.value);
  }

  function handleDelete(event: CustomEvent<{ key: string }>) {
    onDelete(event.detail.key);
  }
</script>

<div class="flex items-center">
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="18px" />
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
</div>

<EditEnvDialog
  bind:open={isEditDialogOpen}
  {keyName}
  {value}
  {existingVariables}
  on:save={handleSave}
/>
<DeleteEnvDialog
  bind:open={isDeleteDialogOpen}
  {keyName}
  on:confirm={handleDelete}
/>
