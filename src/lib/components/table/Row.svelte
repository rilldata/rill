<script>
  export let hovered = false;
  export let row;
  export let width;
</script>

<tr
  style:left={0}
  style:top={0}
  style:transform="translateY({row.start}px)"
  style:height="{row.size}px"
  style:width="{width}px"
  class="absolute hover:bg-gray-100 {hovered ? 'bg-gray-100' : ''}"
>
  <slot />
</tr>
