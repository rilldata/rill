<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
  } from "@rilldata/web-common/components/dialog";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import MemoryManager from "./MemoryManager.svelte";

  export let open = false;
  export let conversationBasePath: string | undefined = undefined;
</script>

<Dialog bind:open>
  <DialogContent class="max-w-3xl">
    <DialogHeader>
      <DialogTitle>{m.chat_memory_title()}</DialogTitle>
    </DialogHeader>
    <div class="max-h-[70vh] overflow-y-auto">
      <MemoryManager {conversationBasePath} />
    </div>
  </DialogContent>
</Dialog>
