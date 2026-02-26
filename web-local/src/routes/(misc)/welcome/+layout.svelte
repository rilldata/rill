<script>
  import { fly } from "svelte/transition";
</script>

<div class="flex size-full overflow-hidden">
  <div class="scroll" in:fly={{ duration: 1600, delay: 400, y: 8 }}>
    <div class="wrapper column p-10 2xl:py-16">
      <slot />
    </div>
  </div>
</div>

<style lang="postcss">
  .scroll {
    @apply size-full overflow-x-hidden overflow-y-auto;
  }

  .wrapper {
    @apply w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .wrapper {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
