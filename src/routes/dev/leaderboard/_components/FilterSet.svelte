<script>
  import { slide } from "svelte/transition";
</script>

<div
  transition:slide={{ duration: 200 }}
  class="grid gap-x-4 items-center"
  style:grid-template-rows="max-content max-content"
>
  <div class="font-bold">
    <slot name="name" />
  </div>
  <div class="flex flex-wrap gap-x-3 gap-y-1">
    <slot name="values" />
  </div>
</div>
