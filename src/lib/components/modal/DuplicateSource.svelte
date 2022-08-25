<script lang="ts">
  import {
    DuplicateActions,
    duplicateSourceAction,
    duplicateSourceName,
  } from "$lib/application-state-stores/application-store";
  import { createEventDispatcher } from "svelte";
  import { Dialog } from "../modal-new";

  const dispatch = createEventDispatcher();

  function onCancel() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Cancel;
    dispatch("cancel");
  }
</script>

<Dialog
  showCancel
  on:cancel={onCancel}
  on:submit={() => {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Overwrite;
  }}
>
  <svelte:fragment slot="title">Duplicate source name</svelte:fragment>
  <svelte:fragment slot="body">
    A source with the name <b>{$duplicateSourceName}</b> already exists.
  </svelte:fragment>
  <svelte:fragment slot="submit-body">Replace existing source</svelte:fragment>
</Dialog>
