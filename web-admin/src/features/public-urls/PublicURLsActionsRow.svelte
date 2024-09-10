<script lang="ts">
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { onDestroy } from "svelte";
  import { CheckIcon, CopyIcon } from "lucide-svelte";

  export let id: string;
  export let url: string;
  export let onDelete: (deletedTokenId: string) => void;

  let copied = false;
  let copyTimer: ReturnType<typeof setTimeout>;

  const COPIED_TIMER = 1_500;

  function handleCopy(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    navigator.clipboard.writeText(url);
    copied = true;

    if (copyTimer) clearTimeout(copyTimer);

    copyTimer = setTimeout(() => {
      copied = false;
    }, COPIED_TIMER);
  }

  async function handleDelete(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    try {
      onDelete(id);
    } catch (error) {
      console.error("Failed to delete magic auth token:", error);
    }
  }

  onDestroy(() => {
    if (copyTimer) clearTimeout(copyTimer);
  });
</script>

<div class="flex items-center justify-center gap-x-4">
  {#if copied}
    <button class="text-gray-400 hover:text-gray-500">
      <CheckIcon size="14px" />
    </button>
  {:else if url}
    <button on:click={handleCopy} class="text-gray-400 hover:text-gray-500">
      <CopyIcon size="14px" />
    </button>
  {:else}
    <!-- Avoid layout shift -->
    <!-- TODO: We can add info icon to handle previously created public urls if we want -->
    <!-- If we add an info icon, when do we remove it ykwim? -->
    <div class="h-[14px] w-[14px]" />
  {/if}

  <button on:click={handleDelete} class="text-gray-400 hover:text-gray-500">
    <Trash size="14px " />
  </button>
</div>
