<script lang="ts">
  import Modal from "$lib/components/modal/Modal.svelte";
  import ModalAction from "$lib/components/modal/ModalAction.svelte";
  import ModalActions from "$lib/components/modal/ModalActions.svelte";
  import ModalContent from "$lib/components/modal/ModalContent.svelte";
  import ModalTitle from "$lib/components/modal/ModalTitle.svelte";

  export let showExportErrorModal = false;
  export let exportErrorMessage = "";
</script>

<Modal
  open={showExportErrorModal}
  onBackdropClick={() => (showExportErrorModal = false)}
>
  <ModalTitle>export error</ModalTitle>
  <ModalContent>{exportErrorMessage}</ModalContent>
  <ModalActions>
    <ModalAction on:click={() => (showExportErrorModal = false)}>
      close
    </ModalAction>
  </ModalActions>
</Modal>
