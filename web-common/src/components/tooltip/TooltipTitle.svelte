<div
  class="grid gap-x-2 pointer-events-none pt-1 pb-1 items-baseline"
  style="grid-template-columns: auto max-content"
  style:min-width="200px"
>
  <div
    class="font-bold truncate text-gray-100 dark:text-gray-200"
    aria-label="tooltip-name"
  >
    <slot name="name" />
  </div>
  <div
    class="text-gray-300 justify-self-end pl-3"
    style:max-width="280px"
    aria-label="tooltip-name-description"
  >
    <slot name="description" />
  </div>
</div>
