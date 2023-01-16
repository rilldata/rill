<script lang="ts">
  export let active = false;
</script>

<button
  on:click
  class:bg-gray-100={active}
  class:bg-gray-50={!active}
  class="w-full rounded flex items-center gap-x-2 justify-center py-1 border hover:bg-gray-100 text-gray-500"
  ><slot /></button
>
