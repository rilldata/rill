<div
  class="grid gap-x-2 items-center pt-1 pb-1 items-baseline"
  style="grid-template-columns: auto max-content"
  style:min-width="200px"
>
  <div class="font-bold truncate">
    <slot name="name" />
  </div>
  <div class="text-gray-300 justify-self-end pl-3" style:max-width="350px">
    <slot name="description" />
  </div>
</div>
