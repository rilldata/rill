<!--  Intercepted navigation follows this example:
https://github.com/sveltejs/kit/pull/3293#issuecomment-1011553037 -->

<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { useIsSourceUnsaved } from "@rilldata/web-common/features/sources/selectors";
  import { emitNavigationTelemetry } from "../../../layout/navigation/navigation-utils";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useSourceStore } from "../sources-store";
  import UnsavedSourceDialog from "./UnsavedSourceDialog.svelte";

  export let sourceName: string;

  const sourceStore = useSourceStore(sourceName);

  let interceptedNavigation: string | null = null;

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML,
  );

  $: isSourceUnsaved = $isSourceUnsavedQuery.data;

  $: file = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
  );

  function handleCancel() {
    interceptedNavigation = null;
  }

  async function handleConfirm() {
    // Revert clientYAML to the last saved version
    sourceStore.set({ clientYAML: $file.data?.blob || "" });

    // Navigate to the new page
    if (interceptedNavigation) {
      emitNavigationTelemetry(interceptedNavigation, "source").catch(
        console.error,
      );
      await goto(interceptedNavigation);
    }

    // Reset the intercepted navigation
    interceptedNavigation = null;
  }

  beforeNavigate(({ to, cancel }) => {
    const toHref = to?.url.href;

    if ((!isSourceUnsaved || interceptedNavigation) && toHref) {
      emitNavigationTelemetry(toHref, "source").catch(console.error);
      return;
    }

    // The current source is unsaved AND the confirmation dialog has not yet been shown
    cancel();

    if (toHref) {
      interceptedNavigation = toHref;
    }
  });
</script>

<slot />

{#if interceptedNavigation}
  <UnsavedSourceDialog on:confirm={handleConfirm} on:cancel={handleCancel} />
{/if}
