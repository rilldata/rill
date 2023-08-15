<script lang="ts">
  import { format as d3format } from "d3-format";
  import type { TCellData } from "./mock-data";
  export let cell: TCellData;
  export let posClass: string = undefined;
  export let negClass: string = undefined;
  export let format = "";

  $: applyFormat = d3format(format);

  let _class = "";
  $: {
    _class = "text-right w-full";
    if (posClass && cell.value > 0) {
      _class += ` ${posClass}`;
    }
    if (negClass && cell.value < 0) {
      _class += ` ${negClass}`;
    }
  }
</script>

<div class={_class}>
  {applyFormat(cell.value)}
</div>
