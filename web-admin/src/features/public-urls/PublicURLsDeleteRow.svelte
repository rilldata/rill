<script lang="ts">
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";

  export let id: string;
  export let onDelete: (deletedTokenId: string) => void;

  async function handleClick(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    try {
      onDelete(id);
    } catch (error) {
      console.error("Failed to delete magic auth token:", error);
    }
  }
</script>

<div class="flex items-center justify-center">
  <button on:click={handleClick} class="text-gray-400 hover:text-gray-500">
    <Trash size="16px " />
  </button>
</div>
