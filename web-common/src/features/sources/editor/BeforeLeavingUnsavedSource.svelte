<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { Button } from "../../../components/button";
  import { Dialog } from "../../../components/modal";
  import DialogFooter from "../../../components/modal/dialog/DialogFooter.svelte";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  // Include `$file.dataUpdatedAt` and `clientYAML` in the reactive statement to recompute
  // the `isSourceUnsaved` value whenever they change
  const file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );
  const sourceStore = useSourceStore();
  $: isSourceUnsaved =
    $file.dataUpdatedAt &&
    $sourceStore.clientYAML &&
    useIsSourceUnsaved($runtime.instanceId, sourceName);

  // Intercepted navigation follows this example:
  // https://github.com/sveltejs/kit/pull/3293#issuecomment-1011553037

  let intercepted = null;

  const handleCancel = () => {
    intercepted = null;
  };

  const handleConfirm = () => {
    goto(intercepted.url);
  };

  beforeNavigate((nav) => {
    if (!isSourceUnsaved) return;
    if (intercepted) return;

    nav.cancel();

    if (nav.to) {
      intercepted = { url: nav.to.url.href };
    }
  });
</script>

<slot />

{#if intercepted}
  <Dialog on:cancel={close} size="sm" useContentForMinSize>
    <svelte:fragment slot="title">Leave source without saving?</svelte:fragment>
    <div class="text-sm text-slate-500" slot="body">
      Navigating away will lose your changes.
    </div>
    <DialogFooter slot="footer">
      <Button type="secondary" on:click={handleCancel}>Keep editing</Button>
      <Button type="primary" on:click={handleConfirm}>Yes, leave source</Button>
    </DialogFooter>
  </Dialog>
{/if}
