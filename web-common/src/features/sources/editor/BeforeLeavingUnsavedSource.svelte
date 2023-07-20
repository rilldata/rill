<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { useIsSourceUnsaved } from "@rilldata/web-common/features/sources/selectors";
  import { emitNavigationTelemetry } from "../../../layout/navigation/navigation-utils";
  import { currentHref } from "../../../layout/navigation/stores";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useSourceStore } from "../sources-store";
  import UnsavedSourceDialog from "./UnsavedSourceDialog.svelte";

  export let sourceName: string;

  const sourceStore = useSourceStore(sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;

  // Intercepted navigation follows this example:
  // https://github.com/sveltejs/kit/pull/3293#issuecomment-1011553037

  let interceptedNavigation = null;

  const handleCancel = () => {
    interceptedNavigation = null;
  };

  const handleConfirm = () => {
    goto(interceptedNavigation.url);
    interceptedNavigation = null;
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
