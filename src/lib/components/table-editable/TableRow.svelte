<script>
  export let hovered = false;
</script>

<tr
  class="
        hover:bg-gray-100
        {hovered && 'bg-gray-100'}
    "
>
  <slot />
</tr>
