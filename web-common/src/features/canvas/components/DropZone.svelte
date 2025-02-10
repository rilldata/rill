<script lang="ts">
  export let column: number;
  export let row: number;
  export let allowDrop: boolean;
  export let maxColumns: number;
  export let onHover: (id: string) => void;
  export let onMouseLeave: () => void;
  export let onDrop: (row: number, column: number) => void;
</script>

{#each { length: 2 } as _, i (i)}
  {@const effectiveColumn = column + i}
  <div
    class:left={i === 0}
    class:first={effectiveColumn === 0}
    class:last={effectiveColumn === maxColumns}
    class:even={effectiveColumn % 2 === 0}
    class:pointer-events-auto={allowDrop}
    style:height="calc(100% - 80px)"
    class="absolute z-20 top-10 h-full"
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
  div {
    @apply opacity-10 w-1/2;
  }

  .left {
    @apply left-0;
  }

  :not(.left) {
    @apply right-0;
    @apply opacity-10;
  }

  div:hover {
    @apply opacity-50;
  }

  .first {
    width: calc(50% + 80px);
    @apply -left-20;
  }

  .last {
    width: calc(50% + 80px);
    @apply -right-20;
  }

  /* .even {
    @apply bg-red-400;
  }

  :not(.even) {
    @apply bg-green-400;
  } */
</style>
