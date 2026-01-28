<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useRuntimeVersion } from "./selectors";

  $: runtimeVersionQuery = useRuntimeVersion();
  $: version = $runtimeVersionQuery.data?.version || "unknown";
</script>

<section class="version-info">
  <h3 class="version-label">Version</h3>
  {#if $runtimeVersionQuery.isLoading}
    <div class="py-1">
      <Spinner status={EntityStatus.Running} size="12px" />
    </div>
  {:else}
    <span class="version-value">{version}</span>
  {/if}
</section>

<style lang="postcss">
  .version-info {
    @apply flex flex-col gap-y-1;
  }

  .version-label {
    @apply text-[10px] leading-none font-semibold uppercase;
    @apply text-fg-secondary;
  }

  .version-value {
    @apply text-fg-primary text-[11px];
  }
</style>
