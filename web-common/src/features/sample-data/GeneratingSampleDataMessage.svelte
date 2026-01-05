<script lang="ts">
  import { SparklesIcon } from "lucide-svelte";

  // row-span-4 row-span-2
  const heights = [4, 2, 2, 2];
  // col-span-1 col-span-3
  const widths = [1, 3, 3, 3];
  const delays = [0, 0.5, 1, 1.5];
</script>

<div class="container">
  <div class="loading-grid">
    {#each heights as height, i (i)}
      {#each widths as width, j (j)}
        <div
          class="row-span-{height} col-span-{width} loading-grid-cell"
          style="animation-delay: {delays[j]}s"
        ></div>
      {/each}
    {/each}
  </div>
  <div class="title">
    <SparklesIcon size="14px" class="text-blue-500" />
    <span>Generating your sample data...</span>
  </div>
  <div class="description">
    Tips: you can ask AI to create and modify your dashboard
  </div>
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col px-8 m-auto items-center w-fit;
  }

  .loading-grid {
    @apply grid grid-cols-10 gap-3 p-8 h-52 w-96 bg-surface;
    @apply border-2 border-primary-200 rounded-lg border-dotted;
  }

  .loading-grid-cell {
    @apply animate-pulse bg-accent rounded-lg;
  }

  .title {
    @apply flex flex-row items-center mt-4 gap-x-1 text-base font-semibold;
  }

  .description {
    @apply text-sm;
  }
</style>
