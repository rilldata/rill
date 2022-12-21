<div
  style:width="300px"
  class="px-2 py-2 long-tooltip-description flex flex-col gap-y-2"
>
  <slot />
</div>

<style lang="postcss">
  /** we are setting the tooltip content code block styling here, since this is sometimes 
programmatically returned by the runtime **/
  :global(.long-tooltip-description code) {
    @apply px-2 py-1 bg-gray-900 text-gray-100 rounded break-words w-max;
    transform: translateX(-0.5rem);
    display: block;
  }
</style>
