<script>
  import AlertCircle from "@rilldata/web-common/components/icons/AlertCircle.svelte";
  import { Dialog, DialogOverlay } from "@rgossiaux/svelte-headlessui";
  import { Button } from "@rilldata/web-common/components/button/index";
  import { createEventDispatcher } from "svelte";

  let open = true;
  const dispatch = createEventDispatcher();

  function handleContinue() {
    open = false;
    dispatch("continue");
  }

  function handleClose() {
    open = false;
    dispatch("close");
  }
</script>

<Dialog class="fixed inset-0 flex items-center justify-center z-50" {open}>
  <DialogOverlay
    class="fixed inset-0 bg-gray-400 transition-opacity opacity-40"
  />
  <div
    class="transform bg-white rounded-md border border-slate-300 flex flex-col gap-2 shadow-lg w-[550px] p-6"
  >
    <div class="flex flex-row gap-3.5">
      <div class="p-0.5"><AlertCircle color="#eab308" size="24px" /></div>
      <div class="flex flex-col gap-0.5">
        <div class="text-lg font-semibold text-slate-900">
          Cancel alert creation?
        </div>
        <div class="text-slate-500 text-sm">
          You haven’t created this alert yet, so closing this window will lose
          your work. To create alert, click Create under Delivery.
        </div>
      </div>
    </div>
    <div class="flex items-center gap-x-2">
      <div class="grow" />
      <Button on:click={handleContinue} type="secondary">Keep Working</Button>
      <Button on:click={handleClose} type="primary">Close</Button>
    </div>
  </div>
</Dialog>
