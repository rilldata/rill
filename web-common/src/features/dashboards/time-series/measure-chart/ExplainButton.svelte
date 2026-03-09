<script lang="ts">
  import { Bot } from "lucide-svelte";
  import type { PlotBounds } from "./types";

  export let x: number;
  export let plotBounds: PlotBounds;
  export let onClick: () => void;
</script>

<button
  class="absolute flex items-center gap-x-1 cursor-pointer hover:underline text-primary text-xs bg-transparent border-none p-0"
  style:left="{x - 35}px"
  style:top="{plotBounds.top + plotBounds.height + 2}px"
  on:click|stopPropagation={onClick}
>
  <Bot size={14} class="stroke-primary -mt-0.5" />
  <span>Explain (E)</span>
</button>
