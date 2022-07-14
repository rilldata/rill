<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let primary = false;
  export let disabled = false;

  const dispatch = createEventDispatcher();

  const handleClick = () => {
    if (!disabled) {
      dispatch("click");
    }
  };
</script>

<button
  {disabled}
  class="px-4 py-2 rounded border border-gray-500 flex flex-row gap-x-2 {primary
    ? 'bg-blue-400/50 hover:bg-blue-300 hover:border-blue-300'
    : 'hover:bg-gray-200 hover:border-gray-200 transition-tranform duration-100'}
    disabled:cursor-not-allowed disabled:bg-gray-200 disabled:border-gray-200 disabled:opacity-50"
  on:click={handleClick}
>
  <slot />
</button>
