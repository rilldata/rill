<script lang="ts">
  export let fields: string[];
  export let selected: number = -1;
  export let onClick: (i: number, value: string) => void = () => {};
</script>

<div class="option-wrapper">
  {#each fields as field, i (field)}
    <button
      on:click={() => onClick(i, field)}
      class="-ml-[1px] first-of-type:-ml-0 px-2 border border-gray-300 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
      class:selected={selected === i}
    >
      {field}
    </button>
  {/each}
</div>

<style lang="postcss">
  button {
    @apply capitalize;
  }

  .option-wrapper {
    @apply flex h-6 text-sm w-fit mb-1 rounded-[2px];
  }

  .option-wrapper > .selected {
    @apply border-primary-500 z-50 text-primary-500;
  }
</style>
