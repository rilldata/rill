<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { useIsSourceUnsaved } from "@rilldata/web-common/features/sources/selectors";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { emitNavigationTelemetry } from "../../../layout/navigation/navigation-utils";
  import { currentHref } from "../../../layout/navigation/stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EntityType } from "../../entity-management/types";
  import { useSourceStore } from "../sources-store";
  import UnsavedSourceDialog from "./UnsavedSourceDialog.svelte";

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

  let interceptedNavigation = null;

  const handleCancel = () => {
    interceptedNavigation = null;
  };

  const handleConfirm = () => {
    goto(interceptedNavigation.url);
  };

  function navigate(href: string) {
    currentHref.set(href);
    emitNavigationTelemetry(href);
  }

  beforeNavigate((nav) => {
    const toHref = nav.to.url.href;

    if (!isSourceUnsaved) {
      navigate(toHref);
      return;
    }
    if (interceptedNavigation) {
      navigate(toHref);
      return;
    }

    // The current source is unsaved AND the confirmation dialog has not yet been shown
    nav.cancel();

    if (nav.to) {
      interceptedNavigation = { url: toHref };
    }
  });
</script>

<slot />

{#if interceptedNavigation}
  <UnsavedSourceDialog on:confirm={handleConfirm} on:cancel={handleCancel} />
{/if}
