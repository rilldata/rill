<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
  } from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { onMount } from "svelte";

  export let value: string | object;
  export let onClose: () => void;

  let isOpen = true;
  let textarea: HTMLTextAreaElement;

  function handleCopy() {
    const textToCopy =
      typeof value === "object" ? JSON.stringify(value, null, 2) : value;
    copyToClipboard(textToCopy, "Cell value copied to clipboard");
  }

  function handleClose() {
    isOpen = false;
    onClose();
  }

  $: formattedValue =
    typeof value === "object" ? JSON.stringify(value, null, 2) : value;
</script>

<Dialog bind:open={isOpen} onOpenChange={handleClose}>
  <DialogContent class="sm:max-w-[600px]">
    <DialogHeader>
      <DialogTitle>Cell Inspector</DialogTitle>
    </DialogHeader>
    <div class="flex flex-col gap-4">
      <div class="flex justify-end">
        <Button type="secondary" on:click={handleCopy}>Copy</Button>
      </div>

      <div class="relative">
        <textarea
          bind:this={textarea}
          class="w-full h-[400px] p-4 rounded-md border bg-background font-mono text-sm"
          readonly
          value={formattedValue}
        />
      </div>
    </div>
  </DialogContent>
</Dialog>

<style>
  textarea {
    resize: none;
    white-space: pre;
    overflow-wrap: normal;
    overflow-x: auto;
  }
</style>
