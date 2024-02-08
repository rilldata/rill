<script lang="ts">
  import {
    Dialog,
    DialogOverlay,
    DialogTitle,
  } from "@rgossiaux/svelte-headlessui";

  export let open: boolean;
  export let titleMarginBottomOverride: string | null = null;
  export let widthOverride: string | null = null;

  let titleMarginBottom = titleMarginBottomOverride ?? "mb-2";
  let width = widthOverride ?? "max-w-lg w-full";
</script>

<Dialog
  {open}
  on:close
  class="fixed inset-0 flex items-center justify-center z-50"
>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <div
    class="transform p-6 bg-white rounded-md border border-slate-300 flex flex-col shadow-lg {width}"
  >
    <DialogTitle
      class="text-gray-900 text-lg font-semibold leading-7 {titleMarginBottom}"
    >
      <slot name="title" />
    </DialogTitle>
    <slot name="body" />
    <slot name="footer" />
  </div>
</Dialog>
