<script lang="ts">
  import type { Side } from "./types";

  export let side: Side = "left";
  export let classes: string = undefined;
</script>

<div
  style:grid-column={side}
  class="{side === 'left' ? 'text-left' : 'text-right'} ${classes ?? ''}"
>
  <slot />
</div>
