<script>
  export let pad = true;
  export let padTop = false;
  export let padBottom = pad;
</script>

<div
  class=" grid gap-x-6 items-baseline text-gray-100"
  class:pb-1={padBottom}
  class:pt-1={padTop}
  style="grid-template-columns: auto max-content"
>
  <slot />
</div>
