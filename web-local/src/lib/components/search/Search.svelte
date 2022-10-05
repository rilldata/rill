<script>
  import { onMount } from "svelte";
  import Search from "../icons/Search.svelte";

  /* Autofocus search bar on mount */
  export let autofocus = true;
  /* Get reference of input DOM element */
  export let ref = null;
  /* Input value being searched */
  export let value;

  onMount(() => {
    if (autofocus) window.requestAnimationFrame(() => ref.focus());
  });
</script>

<form class="flex items-center">
  <div class="relative w-full px-1 mb-1">
    <div class="flex absolute inset-y-0 items-center pl-2">
      <Search />
    </div>
    <input
      bind:this={ref}
      type="text"
      autocomplete="off"
      class="outline-none bg-gray-100 border border-gray-200
        rounded-sm focus:border-gray-300 
        block w-full pl-8 p-1 placeholder:italic"
      placeholder="Search"
      bind:value
      on:input
    />
  </div>
</form>
