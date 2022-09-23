<script lang="ts">
  export let disabled = false;
  export let rounded = false;
  export let compact = false;
</script>

<button
  on:click
  class:cursor-auto={disabled}
  class:rounded
  class="ml-3 grid place-items-center 
{disabled ? 'text-gray-400' : 'text-gray-800'}

{disabled ? '' : 'hover:bg-gray-200'}"
  style:width="{compact ? 20 : 24}px"
  style:height="{compact ? 20 : 24}px"
  style:font-size="18px"
>
  <slot />
</button>
