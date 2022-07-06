<script lang="ts">
  import {
    DuplicateActions,
    duplicateSourceAction,
    duplicateSourceName,
  } from "$lib/application-state-stores/application-store";
  import Modal from "$lib/components/modal/Modal.svelte";
  import ModalAction from "$lib/components/modal/ModalAction.svelte";
  import ModalActions from "$lib/components/modal/ModalActions.svelte";
  import ModalContent from "$lib/components/modal/ModalContent.svelte";
  import ModalTitle from "$lib/components/modal/ModalTitle.svelte";
</script>

<Modal open={$duplicateSourceName !== null} onBackdropClick={() => undefined}>
  <ModalTitle>duplicate source found</ModalTitle>
  <ModalContent
    >a source with the name {$duplicateSourceName} already exists</ModalContent
  >
  <ModalActions>
    <ModalAction
      onClick={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.Overwrite;
      }}
    >
      replace
    </ModalAction>
    <ModalAction
      onClick={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.KeepBoth;
      }}
    >
      keep both
    </ModalAction>
    <ModalAction
      onClick={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.Cancel;
      }}
    >
      cancel
    </ModalAction>
  </ModalActions>
</Modal>
