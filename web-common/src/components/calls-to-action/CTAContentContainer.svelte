<div class="cta-content-container">
  <slot />
</div>

<style lang="postcss">
  .cta-content-container {
    @apply w-full max-w-[570px] text-center;
    @apply flex flex-col justify-center items-center gap-y-6;
  }
</style>
