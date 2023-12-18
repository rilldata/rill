<footer>
  <slot />
</footer>

<style lang="postcss">
  footer {
    @apply flex flex-row mt-auto items-center justify-end;
    @apply bg-slate-100;
    @apply border-t border-slate-300;
    @apply bottom-0 sticky;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
