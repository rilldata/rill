<script lang="ts">
  export let id = "";
  export let label = "";
  export let error: string;
  export let value: string;
</script>

<label for={id} class="text-xs">{label}</label>
<input
  autocomplete="off"
  type="text"
  {id}
  class="border border-gray-400 rounded px-1 py-2 cursor-pointer focus:outline-blue-500 w-full text-xs"
  bind:value
/>
{#if error}
  <div class="text-red-500 text-xs">{error}</div>
{/if}
