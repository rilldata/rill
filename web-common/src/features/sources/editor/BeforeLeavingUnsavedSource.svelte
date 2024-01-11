<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { useIsSourceUnsaved } from "@rilldata/web-common/features/sources/selectors";
  import { emitNavigationTelemetry } from "../../../layout/navigation/navigation-utils";
  import { currentHref } from "../../../layout/navigation/stores";
  import { createRuntimeServiceGetFile } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useSourceStore } from "../sources-store";
  import UnsavedSourceDialog from "./UnsavedSourceDialog.svelte";

  export let sourceName: string;

  const sourceStore = useSourceStore(sourceName);

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

  // Intercepted navigation follows this example:
  // https://github.com/sveltejs/kit/pull/3293#issuecomment-1011553037

  let interceptedNavigation: { url: string | URL } | null = null;

  const handleCancel = () => {
    interceptedNavigation = null;
  };

  const handleConfirm = () => {
    // Revert clientYAML to the last saved version
    sourceStore.set({ clientYAML: $file.data?.blob || "" });

    // Navigate to the new page
    if (interceptedNavigation?.url) {
      goto(interceptedNavigation.url);
    }

    // Reset the intercepted navigation
    interceptedNavigation = null;
  };

  function navigate(href: string) {
    currentHref.set(href);
    emitNavigationTelemetry(href);
  }

  beforeNavigate((nav) => {
    const toHref = nav?.to?.url.href;

    if (!isSourceUnsaved && toHref) {
      navigate(toHref);
      return;
    }
    if (interceptedNavigation && toHref) {
      navigate(toHref);
      return;
    }

    // The current source is unsaved AND the confirmation dialog has not yet been shown
    nav.cancel();

    if (nav.to && toHref) {
      interceptedNavigation = { url: toHref };
    }
  });
</script>

<slot />

{#if interceptedNavigation}
  <UnsavedSourceDialog on:confirm={handleConfirm} on:cancel={handleCancel} />
{/if}
