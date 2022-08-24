<script>
  import IconButton from "$lib/components/button/IconButton.svelte";

  import Close from "$lib/components/icons/Close.svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import ModalContainer from "../ModalContainer.svelte";
  import DialogHeader from "./DialogHeader.svelte";
  export let dark = false;
  export let showCancel = true;

  const dispatch = createEventDispatcher();

  $: containerClasses = dark
    ? "text-white bg-gray-800"
    : "text-gray-800 bg-white";

  $: cancelButtonClasses = dark ? "hover:bg-gray-400" : "hover:bg-gray-200";
  let mounted = false;
  onMount(() => {
    mounted = true;
  });
</script>

<ModalContainer on:cancel>
  <div class="grid place-items-center  w-screen h-screen">
    <div
      class="w-96 {containerClasses} rounded"
      style:transform="translateY(-120px)"
    >
      <DialogHeader>
        <svelte:fragment slot="title"><slot name="title" /></svelte:fragment>
        <div slot="right">
          {#if showCancel}
            <!-- FIXME: this should be replaced with the IconButton in an open PR -->
            <IconButton
              on:click={() => {
                dispatch("cancel");
              }}
            >
              <Close size="16px" />
            </IconButton>
          {:else}
            <slot name="right" />
          {/if}
        </div>
      </DialogHeader>
      <hr />
      <div class="px-7 pt-8 pb-16">
        <slot name="body" />
      </div>
      <footer class="p-2">
        <slot name="footer" />
      </footer>
    </div>
  </div>
</ModalContainer>
