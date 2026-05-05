<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { V1ParseError } from "@rilldata/web-common/runtime-client";

  export let parseErrors: V1ParseError[] = [];
  export let parserReconcileError: string | undefined = undefined;
  export let isLoading = false;
  export let isError = false;
  export let errorMessage = "";
</script>

<section class="flex flex-col gap-y-4">
  <h3 class="parse-errors-header">
    Parse Errors
    {#if parseErrors.length > 0}
      <span class="parse-errors-badge">{parseErrors.length}</span>
    {/if}
  </h3>

  {#if isLoading}
    <Spinner status={EntityStatus.Running} size={"16px"} />
  {:else if isError}
    <div class="text-red-500">
      Error loading parse errors: {errorMessage}
    </div>
  {:else if parseErrors.length > 0}
    <div class="parse-errors-list">
      {#each parseErrors as error ((error.filePath ?? "") + ":" + error.message)}
        <div class="parse-error-item">
          {#if error.filePath}
            <span class="parse-error-file">{error.filePath}</span>
          {/if}
          <span class="parse-error-message">{error.message}</span>
        </div>
      {/each}
    </div>
  {:else if parserReconcileError}
    <div class="text-red-500">
      {parserReconcileError}
    </div>
  {:else}
    <p class="text-sm text-fg-secondary">No parse errors</p>
  {/if}
</section>

<style lang="postcss">
  .parse-errors-header {
    @apply text-sm font-semibold text-fg-primary flex items-center gap-2;
  }
  .parse-errors-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .parse-errors-list {
    @apply flex flex-col gap-2;
  }
  .parse-error-item {
    @apply flex flex-col gap-0.5 px-3 py-2 rounded-md bg-red-50 text-sm;
  }
  .parse-error-file {
    @apply font-mono text-xs text-fg-secondary;
  }
  .parse-error-message {
    @apply text-red-700;
  }
</style>
