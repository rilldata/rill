<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let type: "primary" | "secondary" | "text";
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
  class="px-4 py-2 rounded flex flex-row gap-x-2 transition-transform duration-100
    {type === 'primary' &&
    'bg-gray-800 border border-gray-800 hover:bg-gray-900 hover:border-gray-900 text-gray-100 hover:text-white'}
    {type === 'secondary' &&
    'border border-gray-500 hover:bg-gray-200 hover:border-gray-200'}
    {type === 'text' && 'text-gray-900 hover:bg-gray-300'}
    disabled:cursor-not-allowed disabled:text-gray-700 disabled:bg-gray-200 disabled:border disabled:border-gray-400 disabled:opacity-50"
  on:click={handleClick}
>
  <slot />
</button>
