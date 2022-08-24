<script lang="ts">
  import {
    DuplicateActions,
    duplicateSourceAction,
    duplicateSourceName,
  } from "$lib/application-state-stores/application-store";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../button";
  import { Dialog } from "../modal-new";

  const dispatch = createEventDispatcher();

  function onCancel() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Cancel;
    dispatch("cancel");
  }
</script>

<Dialog showCancel on:cancel={onCancel}>
  <svelte:fragment slot="title">< Duplicate source name</svelte:fragment>
  <svelte:fragment slot="body">
    A source with the name <b>{$duplicateSourceName}</b> already exists.
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <Button on:click={onCancel} type="text">Cancel</Button>
    <Button
      on:click={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.Overwrite;
      }}
      status="error"
      type="primary">Replace existing source</Button
    >
  </svelte:fragment>
</Dialog>
<!-- 
<Modal open={$duplicateSourceName !== null} onBackdropClick={() => undefined}>
  <ModalTitle>Duplicate Source Found</ModalTitle>
  <ModalContent
    >A source with the name <b>{$duplicateSourceName}</b> already exists</ModalContent
  >
  <ModalActions>
    <ModalAction
      on:click={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.Overwrite;
      }}
    >
      replace
    </ModalAction>
    <ModalAction
      on:click={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.KeepBoth;
      }}
    >
      keep both
    </ModalAction>
    <ModalAction
      on:click={() => {
        $duplicateSourceName = null;
        $duplicateSourceAction = DuplicateActions.Cancel;
      }}
    >
      cancel
    </ModalAction>
  </ModalActions>
</Modal> -->
