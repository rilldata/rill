<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  export let open = false;
  export let groupName: string;
  export let onRename: (groupName: string, newName: string) => void;

  let newName: string;

  function onNewNameInput(e: any) {
    newName = e.target.value;
  }

  async function handleRename() {
    try {
      onRename(groupName, newName);
      open = false;
    } catch (error) {
      console.error("Failed to rename user group:", error);
    }
  }
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Rename user group</DialogTitle>
    </DialogHeader>
    <DialogFooter class="mt-4">
      <div class="flex flex-col gap-2 w-full">
        <Input
          bind:value={newName}
          placeholder="New name"
          on:input={onNewNameInput}
        />
        <Button type="primary" large on:click={handleRename}>Rename</Button>
      </div>
    </DialogFooter>
  </DialogContent>
</Dialog>
