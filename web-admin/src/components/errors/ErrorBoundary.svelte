<script lang="ts">
  /**
   * This file allows us to navigate to the `ErrorPage` when errors occur during runtime.
   * SvelteKit's `+error.svelte` file is itself an error boundary, but it only gets hit for errors during routing.
   */

  import { afterNavigate } from "$app/navigation";
  import { errorStore, isErrorStoreEmpty } from "./error-store";
  import ErrorPage from "./ErrorPage.svelte";

  afterNavigate(() => {
    // Checks to see if we're on the error page (and navigating away)
    if (!$isErrorStoreEmpty) {
      errorStore.reset();
    }
  });
</script>

{#if $isErrorStoreEmpty}
  <slot />
{:else}
  <ErrorPage
    statusCode={$errorStore.statusCode}
    header={$errorStore.header}
    body={$errorStore.body}
  />
{/if}
