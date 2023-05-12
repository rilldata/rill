<main>
  <h1 class="text-xl mb-12">Design System</h1>
  <section>
    <h2>Surfaces</h2>
    <div class="flex flex-col gap-y-4">
      <div
        class="ui-copy surface-bg swatch dark:border-gray-200 border-gray-500"
      >
        <code class="font-bold">surface-bg</code>
        light mode
      </div>

      <div class="dark">
        <div
          class="ui-copy surface-bg swatch dark:border-gray-200 border-gray-500"
        >
          <code class="font-bold">dark surface-bg</code>
          dark mode
        </div>
      </div>
    </div>
  </section>

  <section>
    <h2>UI Color Classes</h2>
    <div>
      {#each [".ui-copy", ".ui-copy-muted", ".ui-copy-inactive", ".ui-copy-disabled"] as cl}
        <div style:font-size="24px" class={cl.slice(1)}>{cl}</div>
      {/each}
    </div>
    <div class="dark">
      <div class="surface">
        {#each [".ui-copy", ".ui-copy-muted", ".ui-copy-inactive", ".ui-copy-disabled"] as cl}
          <div style:font-size="24px" class={cl.slice(1)}>{cl} (dark)</div>
        {/each}
      </div>
    </div>
  </section>
</main>

<style lang="postcss">
  section {
    margin-bottom: 32px;
  }
  .swatch {
    width: 200px;
    height: 100px;
    @apply rounded p-4 border;
  }
</style>
