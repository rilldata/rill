<script lang="ts">
  export let active = false;
</script>

<button
  on:click
  class:bg-gray-300={active}
  class:text-gray-800={active}
  class:bg-transparent={!active}
  class:text-gray-500={!active}
  class="rounded px-3 py-1"
>
  <slot />
</button>
