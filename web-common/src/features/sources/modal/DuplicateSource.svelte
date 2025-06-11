<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { onDestroy } from "svelte";
  import {
    DuplicateActions,
    duplicateSourceAction,
    duplicateSourceName,
  } from "../sources-store";

  export let onComplete: () => void;
  export let onCancel: () => void;

  function cancel() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Cancel;
    onCancel();
  }

  function keepBoth() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.KeepBoth;
    onComplete();
  }

  function overwriteSource() {
    $duplicateSourceName = null;
    $duplicateSourceAction = DuplicateActions.Overwrite;
    onComplete();
  }

  onDestroy(() => {
    $duplicateSourceName = null;
  });
</script>

<Dialog.Description>
  A source with the name <b>{$duplicateSourceName}</b> already exists.
</Dialog.Description>

<Dialog.Footer>
  <Button type="text" onClick={cancel}>Cancel</Button>

  <Button type="text" onClick={overwriteSource}>Replace Existing Source</Button>

  <Button type="primary" onClick={keepBoth}>Keep Both</Button>
</Dialog.Footer>
