<script lang="ts">
  import Overlay from "./Overlay.svelte";
  import { onSourceDrop } from "$lib/util/file-upload";

  export let showDropOverlay: boolean;
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    class="w-screen h-screen grid place-content-center"
    on:drop|preventDefault|stopPropagation={(e) => {
      showDropOverlay = false;
      onSourceDrop(e);
    }}
    on:mouseup|preventDefault|stopPropagation={() => {
      showDropOverlay = false;
    }}
    on:drag|preventDefault|stopPropagation
    on:dragenter|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
  >
    <div
      class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all"
      style:font-size="32px"
    >
      <span class="place-content-center">
        Drop your files to add new source
      </span>
    </div>
  </div>
</Overlay>
