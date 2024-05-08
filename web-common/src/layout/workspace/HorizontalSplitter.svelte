<div class="wrapper">
  <div class="line" />
  <span class="handle" />
</div>

<style lang="postcss">
  .wrapper {
    @apply flex items-center justify-center;
    @apply w-full h-2 pr-2;
    /* @apply border-2 border-red-400; */
  }

  .handle {
    @apply absolute;
    @apply border-gray-400 border bg-white;
    @apply rounded h-1 w-8;
  }

  .line {
    @apply h-[1px] w-full;
    @apply bg-gray-300;
  }
</style>
