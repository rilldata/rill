<script lang="ts">
  import Modal from "$lib/components/modal/Modal.svelte";
  import ModalAction from "$lib/components/modal/ModalAction.svelte";
  import ModalActions from "$lib/components/modal/ModalActions.svelte";
  import ModalContent from "$lib/components/modal/ModalContent.svelte";
  import ModalTitle from "$lib/components/modal/ModalTitle.svelte";

  export let generateCallback;
  export let open = false;
  export let closeModal;
</script>

<Modal {open} onBackdropClick={() => undefined}>
  <ModalTitle>Replace current metrics?</ModalTitle>
  <ModalContent
    >Continuing will replace all existing measures and dimensions.
  </ModalContent>
  <ModalActions>
    <ModalAction on:click={generateCallback}>continue</ModalAction>
    <ModalAction on:click={closeModal}>cancel</ModalAction>
  </ModalActions>
</Modal>
