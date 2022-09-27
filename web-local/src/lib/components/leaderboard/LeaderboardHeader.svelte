<script>
  export let isActive = false;
</script>

<button
  style:height="32px"
  style:grid-template-columns="auto max-content"
  class:font-semibold={!isActive}
  class="
        pl-2 pr-2
        grid justify-between items-center
        w-full
        border-b
        border-gray-200
        rounded-t
        bg-white
        text-gray-600
    "
  on:click
>
  <slot name="title" />
  <slot name="right" />
</button>
