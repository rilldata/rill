<div class="flex size-full overflow-x-hidden overflow-y-auto">
  <div class="background">
    <div class="h-fit pt-48">
      <slot />
    </div>
  </div>
</div>

<style lang="postcss">
  .background {
    @apply flex flex-col w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .background {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }
</style>
