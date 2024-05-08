<script lang="ts">
  import DialogFooter from "@rilldata/web-common/components/modal/dialog/DialogFooter.svelte";
  import DialogCTA from "@rilldata/web-common/components/modal/dialog/DialogCTA.svelte";
  import { createEventDispatcher, onDestroy } from "svelte";
  import {
    DuplicateActions,
    duplicateSourceAction,
    duplicateSourceName,
  } from "../sources-store";

  const dispatch = createEventDispatcher();

  function onCancel() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Cancel;
    dispatch("cancel");
  }

  function onPrimaryAction() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.KeepBoth;
    dispatch("complete");
  }

  function onSecondaryAction() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Overwrite;
    dispatch("complete");
  }

  onDestroy(() => {
    $duplicateSourceName = null;
  });
</script>

<p class="py-2">
  A source with the name <b>{$duplicateSourceName}</b> already exists.
</p>

<DialogFooter>
  <DialogCTA
    on:cancel={onCancel}
    on:primary-action={onPrimaryAction}
    on:secondary-action={onSecondaryAction}
    showSecondary
  >
    <svelte:fragment slot="secondary-action-body"
      ><slot name="secondary-action-body" />Replace Existing Source</svelte:fragment
    >
    <svelte:fragment slot="primary-action-body"
      ><slot name="primary-action-body" />Keep Both</svelte:fragment
    >
  </DialogCTA>
</DialogFooter>
