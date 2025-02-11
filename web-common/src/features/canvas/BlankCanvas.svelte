<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";

  function handleButtonClick(event: MouseEvent) {
    const contextMenuEvent = new MouseEvent("contextmenu", {
      bubbles: true,
      clientX: event.clientX,
      clientY: event.clientY,
    });
    event.target?.dispatchEvent(contextMenuEvent);
  }
</script>

<div class="size-full p-4 bg-white">
  <button
    type="button"
    class="blank-canvas-button flex flex-col items-center gap-2 p-8 rounded-[6px] border border-slate-200 w-full"
    on:click={handleButtonClick}
  >
    <PlusCircleIcon class="w-6 h-6 text-slate-500" />
    <span class="text-sm font-medium text-slate-500">Add a component</span>
  </button>
</div>

<style lang="postcss">
  .blank-canvas-button:hover {
    /* card-hover */
    box-shadow:
      0px 2px 3px 0px rgba(15, 23, 42, 0.03),
      0px 1px 3px 0px rgba(15, 23, 42, 0.04),
      0px 0px 0px 1px rgba(15, 23, 42, 0.06),
      0px 4px 6px 0px rgba(15, 23, 42, 0.09);
  }
</style>
