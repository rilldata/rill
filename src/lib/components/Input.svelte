<script lang="ts">
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";

  export let id = "";
  export let label = "";
  export let error: string;
  export let value: string;
  export let claimFocusOnMount = false;
  let inputElement;
  if (claimFocusOnMount) {
    onMount(() => {
      inputElement.focus();
    });
  }
</script>

<label for={id} class="text-gray-600 pl-1 pb-1 block">{label}</label>
<input
  bind:this={inputElement}
  autocomplete="off"
  type="text"
  {id}
  class="border border-gray-400 rounded px-1 py-2 cursor-pointer focus:outline-blue-500 w-full text-xs"
  bind:value
/>
{#if error}
  <div
    in:slide|local={{ duration: 200 }}
    class="pl-1 text-red-500 text-xs pt-1"
  >
    {error}
  </div>
{/if}
