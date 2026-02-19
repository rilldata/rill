<script lang="ts">
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { pluralizeKind } from "@rilldata/web-common/features/resources/overview-utils";
  import type { ResourceCount } from "@rilldata/web-common/features/resources/overview-utils";

  export let parseErrorCount: number;
  export let errorsByKind: ResourceCount[];
  export let totalErrors: number;
  export let isLoading: boolean;
  export let isError: boolean;
  export let onSectionClick: () => void;
  export let onParseErrorChipClick: (() => void) | undefined = undefined;
  export let onKindChipClick: ((kind: string) => void) | undefined = undefined;
</script>

{#if totalErrors > 0}
  <button
    class="section section-error section-clickable"
    on:click={onSectionClick}
  >
    <div class="section-header">
      <h3 class="section-title flex items-center gap-2">
        Errors
        <span class="error-badge">{totalErrors}</span>
      </h3>
    </div>
    <div class="error-chips">
      {#if parseErrorCount > 0}
        {#if onParseErrorChipClick}
          <button
            class="error-chip"
            on:click|stopPropagation={onParseErrorChipClick}
          >
            <AlertCircleOutline size="12px" />
            <span class="font-medium">{parseErrorCount}</span>
            <span>Parse error{parseErrorCount !== 1 ? "s" : ""}</span>
          </button>
        {:else}
          <span class="error-chip">
            <AlertCircleOutline size="12px" />
            <span class="font-medium">{parseErrorCount}</span>
            <span>Parse error{parseErrorCount !== 1 ? "s" : ""}</span>
          </span>
        {/if}
      {/if}
      {#each errorsByKind as { kind, label, count } (kind)}
        {#if onKindChipClick}
          <button
            class="error-chip"
            on:click|stopPropagation={() => onKindChipClick?.(kind)}
          >
            {#if resourceIconMapping[kind]}
              <svelte:component this={resourceIconMapping[kind]} size="12px" />
            {/if}
            <span class="font-medium">{count}</span>
            <span>{pluralizeKind(label, count)}</span>
          </button>
        {:else}
          <span class="error-chip">
            {#if resourceIconMapping[kind]}
              <svelte:component this={resourceIconMapping[kind]} size="12px" />
            {/if}
            <span class="font-medium">{count}</span>
            <span>{pluralizeKind(label, count)}</span>
          </span>
        {/if}
      {/each}
    </div>
  </button>
{:else}
  <div class="section">
    <div class="section-header">
      <h3 class="section-title flex items-center gap-2">Errors</h3>
    </div>
    {#if isError}
      <p class="text-sm text-fg-secondary">Unable to check for errors.</p>
    {:else if isLoading}
      <p class="text-sm text-fg-secondary">Checking for errors...</p>
    {:else}
      <p class="text-sm text-fg-secondary">No errors detected.</p>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .section {
    @apply block border border-border rounded-lg p-5 text-left w-full;
  }
  .section-clickable {
    @apply cursor-pointer;
  }
  .section-error {
    @apply border-red-500;
  }
  .section-clickable:hover {
    @apply border-red-600;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .error-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .error-chips {
    @apply flex flex-wrap gap-2;
  }
  .error-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md;
    @apply border border-red-300 bg-red-50 text-red-700;
  }
  button.error-chip:hover {
    @apply border-red-500 bg-red-100;
  }
</style>
