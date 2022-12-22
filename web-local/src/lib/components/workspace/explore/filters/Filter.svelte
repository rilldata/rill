<script>
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
</script>

<div
  class="
  flex flex-row gap-x-1 items-center
  "
  style:border-radius="2rem"
>
  <slot />
  <button
    style:margin-right="-.2rem"
    class="hover:bg-red-100 p-1 rounded-xl "
    on:click
  >
    <Close />
  </button>
</div>
