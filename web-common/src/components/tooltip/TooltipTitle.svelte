<div class="w-full justify-between flex py-1" style:min-width="200px">
  <div class="font-bold truncate w-fit" aria-label="tooltip-name">
    <slot name="name" />
  </div>
  <div
    class="text-gray-300 w-fit"
    style:max-width="280px"
    aria-label="tooltip-name-description"
  >
    <slot name="description" />
  </div>
</div>
