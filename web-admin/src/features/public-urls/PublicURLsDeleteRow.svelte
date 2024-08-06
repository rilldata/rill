<script lang="ts">
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";

  export let url: string;

  function handleClick(e: PointerEvent) {
    e.preventDefault();
    e.stopPropagation();

    console.log("deleting row", url);
  }
</script>

<div class="flex items-center justify-center">
  <button on:click={handleClick} class="text-gray-400 hover:text-gray-500">
    <Trash size="16px " />
  </button>
</div>
