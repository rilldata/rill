<script lang="ts">
  import type { HeaderPosition } from "../types";

  export let header;
  export let position: HeaderPosition = "top";

  let positionClasses;
  $: {
    if (position === "top") {
      positionClasses = "absolute left-0 top-0";
    } else if (position === "left") {
      positionClasses = "absolute left-0";
    } else if (position === "top-left") {
      positionClasses = "sticky left-0 top-0 z-40";
    }
  }
</script>

<div
  style:transform="translateX({header.start}px)"
  style:width="{header.size}px"
  style:height="36px"
  class="{positionClasses}
   bg-white text-left border-b border-b-4 border-r border-r-1"
>
  <div
    class="
    text-ellipsis overflow-hidden whitespace-nowrap
  px-4
  border
  border-gray-200
  border-t-0
  border-l-0
  bg-gray-100
  {position === 'top' && 'py-2 text-left'}
  {position === 'left' && 'py-2'}
  {position === 'top-left' && 'py-2 text-center'}
"
  >
    <slot />
  </div>
</div>
