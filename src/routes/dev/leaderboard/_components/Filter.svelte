<script>
  import Close from "$lib/components/icons/Close.svelte";
</script>

<div
  class="
  bg-slate-200
  p-1 pl-2 pr-2
  flex flex-row gap-x-2 items-center
  "
  style:border-radius="2rem"
>
  <slot />
  <button
    style:margin-right="-.2rem"
    class="hover:bg-slate-300 p-1 rounded-xl "
    on:click
  >
    <Close />
  </button>
</div>
