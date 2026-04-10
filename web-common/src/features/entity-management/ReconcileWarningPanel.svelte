<script lang="ts">
  import AlertTriangle from "@rilldata/web-common/components/icons/AlertTriangle.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { slide } from "svelte/transition";

  export let fileArtifact: FileArtifact;

  let expanded = true;

  $: allWarnings = fileArtifact.getAllWarnings(queryClient);
</script>

{#if $allWarnings.length > 0}
  <div
    transition:slide={{ duration: 200 }}
    class="border-l-4 border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20 text-fg-primary flex flex-col flex-none"
    aria-label="Reconcile warnings"
  >
    <button
      class="flex items-center gap-1.5 w-full px-3 py-2 text-left text-yellow-700 dark:text-yellow-400 hover:bg-yellow-100 dark:hover:bg-yellow-900/30 transition-colors"
      on:click={() => (expanded = !expanded)}
    >
      <AlertTriangle size="14px" />
      <span class="text-xs font-medium flex-1">
        {$allWarnings.length}
        {$allWarnings.length === 1 ? "warning" : "warnings"}
      </span>
      <CaretDownIcon
        size="12px"
        className={expanded ? "" : "transform -rotate-90"}
      />
    </button>
    {#if expanded}
      <div
        transition:slide={{ duration: 150 }}
        class="px-3 pb-3 flex flex-col gap-1.5 max-h-48 overflow-auto"
      >
        {#each $allWarnings as warning (warning.message)}
          <p class="text-xs text-yellow-800 dark:text-yellow-300">
            {warning.message ?? ""}
          </p>
        {/each}
      </div>
    {/if}
  </div>
{/if}
