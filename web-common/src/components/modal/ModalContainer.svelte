<script lang="ts">
  import { createEventDispatcher, onDestroy } from "svelte";
  import { fly } from "svelte/transition";
  import Modal from "./modal";
  import Overlay from "./Overlay.svelte";
  import { lockBodyScrolling, unlockBodyScrolling } from "./scroll";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  export let focusTriggerOnClose = true;

  const dispatch = createEventDispatcher();

  let modal;
  let container;

  let originalTrigger;
  let autoFocusTarget;

  let containerMountedInPortal = false;

  /** post-mount, and post-portal (which is to say, as soon as container is actually mounted)
   * let's go ahead and instantiate the modal.
   */
  function initiateOnMount(containerElement) {
    containerMountedInPortal = true;
    originalTrigger = document.activeElement as HTMLElement;
    modal = new Modal(containerElement);
    lockBodyScrolling(containerElement);

    modal.activate();
    // When the dialog is shown, Safari will attempt to set focus on whatever element has autofocus. This can cause
    // the dialogs's animation to jitter (if it starts offscreen), so we'll temporarily remove the attribute, call
    // `focus({ preventScroll: true })` ourselves, and add the attribute back afterwards.
    //
    autoFocusTarget = document.querySelector("[autofocus]");
    if (autoFocusTarget) {
      autoFocusTarget.removeAttribute("autofocus");
    }
    requestAnimationFrame(() => {
      // Set focus to the autofocus target and restore the attribute
      if (autoFocusTarget) {
        (autoFocusTarget as HTMLInputElement).focus({ preventScroll: true });
      } else {
        container?.focus({ preventScroll: true });
      }

      // Restore the autofocus attribute
      if (autoFocusTarget) {
        autoFocusTarget.setAttribute("autofocus", "");
      }
    });
  }

  $: if (Modal && !containerMountedInPortal && container)
    initiateOnMount(container);

  onDestroy(() => {
    if (!modal) return;
    modal.deactivate();
    unlockBodyScrolling(container);
    if (focusTriggerOnClose && typeof originalTrigger?.focus === "function") {
      setTimeout(() => originalTrigger.focus());
    }
  });

  function handleKeydown(event) {
    const key = event.key;
    if (key === "Escape") {
      dispatch("cancel");
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- <Portal> -->

<div
  use:portal
  bind:this={container}
  class="fixed top-0 left-0 right-0 bottom-0 z-50"
  transition:fly|global={{ duration: 125, y: 4 }}
>
  <Overlay />
  <slot />
</div>

<!-- </Portal> -->

<style>
  :global(.sl-scroll-lock) {
    overflow: hidden !important;
  }
</style>
