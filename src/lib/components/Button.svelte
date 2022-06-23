<script lang="ts">
  export let onClick: (event: any) => Promise<void> | void;
  export let primary: boolean = false;
</script>

<button
  class="px-4 py-2 rounded border border-gray-500 flex flex-row gap-x-2 {primary
    ? 'bg-blue-400/50 hover:bg-blue-300 hover:border-blue-300'
    : 'hover:bg-gray-200 hover:border-gray-200 transition-tranform duration-100'}"
  on:click={onClick}
>
  <slot />
</button>
