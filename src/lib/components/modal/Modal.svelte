<script lang="ts">
  import Portal from "../Portal.svelte";

  export let open: boolean = false;
  export let onBackdropClick: () => void = () => {};
</script>

{#if open}
  <Portal>
    <div
      id="overlay"
      class="fixed z-40 top-0 bottom-0 left-0 right-0 backdrop-contrast-50"
      on:click={onBackdropClick}
    />
    <div
      id="modal"
      class="fixed z-50 w-1/2 lg:w-2/5 xl:w-1/3 2xl:w-1/4 top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 border border-gray-500 rounded px-8 py-6 bg-white"
    >
      <slot />
    </div>
  </Portal>
{/if}
