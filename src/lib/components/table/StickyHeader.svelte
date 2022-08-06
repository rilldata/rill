<script lang="ts">
  export let header;
  export let position: "top" | "left" | "top-left" = "top";
</script>

<th
  style:transform="translateX({header.start}px)"
  style:width="{header.size}px"
  style:height="36px"
  class="
  {position === 'top' && 'top-0'}
  {position === 'left' && 'left-0'}
  {position === 'top-left' && 'top-0 left-0 z-10'}
  absolute bg-white text-left border-b border-b-4 border-r border-r-1  "
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
  {position === 'top-left' && 'py-3'}
"
  >
    <slot />
  </div>
</th>
