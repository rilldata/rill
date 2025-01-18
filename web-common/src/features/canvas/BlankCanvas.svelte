<script lang="ts">
  import { PlusCircleIcon } from "lucide-svelte";
</script>

<div class="size-full p-2 bg-white">
  <!-- FIXME -->
  <button
    type="button"
    class="flex flex-col items-center gap-2 p-8 rounded-[6px] border border-slate-200 w-full"
    on:click={() => console.log("clicked")}
  >
    <PlusCircleIcon class="w-6 h-6 text-slate-500" />
    <span class="text-sm font-medium text-slate-500">Add a component</span>
  </button>
</div>
