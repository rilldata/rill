<script lang="ts">
  import { adminServiceGetEmbeddedAnalytics } from "@rilldata/web-admin/client/gen/default/default";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { onMount } from "svelte";

  export let org: string;
  export let resource: string;
  export let height = "500px";

  let iframeSrc: string | undefined;
  let loading = true;
  let error: string | undefined;

  // credentialless is not standard and throws lint error, but it works on chrome and safari
  const iframeProps = {
    credentialless: true,
  };

  onMount(async () => {
    try {
      const resp = await adminServiceGetEmbeddedAnalytics(org, { resource });
      iframeSrc = resp.iframeSrc;
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load embed";
    } finally {
      loading = false;
    }
  });

  let iframeLoading = true;
</script>

{#if loading}
  <div class="embed-loading">
    <Spinner status={EntityStatus.Running} size="16px" />
  </div>
{:else if error}
  <div class="embed-error">
    <p>{error}</p>
  </div>
{:else if iframeSrc}
  {#if iframeLoading}
    <div class="embed-loading">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {/if}
  <iframe
    {...iframeProps}
    src={iframeSrc}
    title="Rill Embed - {resource}"
    class="embed-iframe"
    style:height
    on:load={() => (iframeLoading = false)}
  />
{/if}

<style lang="postcss">
  .embed-loading {
    @apply flex items-center justify-center py-8;
  }
  .embed-error {
    @apply flex items-center justify-center py-8 text-sm text-red-600;
  }
  .embed-iframe {
    @apply w-full border-none rounded-md;
  }
</style>
