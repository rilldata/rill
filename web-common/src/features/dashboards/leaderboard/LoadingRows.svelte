<script lang="ts">
  export let rows = 7;
  export let columns = 4;
</script>

{#each { length: rows } as _, i (i)}
  <tr>
    <td />
    {#each { length: columns } as _, i (i)}
      <td>
        <div />
      </td>
    {/each}
  </tr>
{/each}

<style lang="postcss">
  td {
    height: 22px;
    @apply p-1 py-[5px];
  }
  div {
    @apply size-full bg-gray-200 animate-pulse rounded-full;
  }
</style>
