{#each { length: 7 } as _, i (i)}
  <tr>
    <td />
    {#each { length: 2 } as _, i (i)}
      <td>
        <div class="loading-bar" />
      </td>
    {/each}
  </tr>
{/each}

<style lang="postcss">
  td {
    height: 22px;
  }
  .loading-bar {
    @apply w-11/12 h-2.5 bg-gray-100 animate-pulse rounded-full;
  }
</style>
