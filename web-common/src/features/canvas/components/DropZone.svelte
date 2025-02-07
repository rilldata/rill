<script lang="ts">
  export let column: number;
  export let row: number;
  export let allowDrop: boolean;
  export let onHover: (id: string) => void;
  export let onMouseLeave: () => void;
  export let onDrop: (row: number, column: number) => void;
</script>

{#each { length: 2 } as _, i (i)}
  {@const effectiveColumn = column + i}
  <div
    class:even={effectiveColumn % 2 === 0}
    class:pointer-events-auto={allowDrop}
    style:height="calc(100% - 80px)"
    class="w-1/2 absolute z-20 top-10 h-full"
    role="presentation"
    on:mouseenter={() => onHover(`${row}-${effectiveColumn}`)}
    on:mouseleave={onMouseLeave}
    on:mouseup={() => {
      if (allowDrop) {
        onDrop(row, effectiveColumn);
      }
    }}
  />
{/each}

<style lang="postcss">
  /* .even {
    @apply left-0;
    @apply bg-red-400;
    @apply opacity-10;
  }

  :not(.even) {
    @apply right-0;
    @apply bg-blue-400;
    @apply opacity-10;
  }

  div:hover {
    @apply opacity-50;
  } */
</style>
