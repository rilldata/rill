<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    registerErrorHandler,
    getUserErrorMessage,
    getRecoveryAction,
    type ErrorContext,
    ErrorSeverity,
  } from "./error-handling";
  import type { ComponentType, SvelteComponentTyped } from "svelte";
  type FallbackComponent = ComponentType<
    SvelteComponentTyped<{ error: Error | null }>
  >;
  export let fallback: FallbackComponent | null = null;
  export let onError: ((error: Error, context: ErrorContext) => void) | null =
    null;

  let error: Error | null = null;
  let errorContext: ErrorContext | null = null;
  let unregisterHandler: (() => void) | null = null;

  onMount(() => {
    // Register error handler
    unregisterHandler = registerErrorHandler((err, ctx) => {
      if (ctx.showToUser) {
        error = err;
        errorContext = ctx;
        onError?.(err, ctx);
      }
    });
  });

  onDestroy(() => {
    unregisterHandler?.();
  });

  function handleReset() {
    error = null;
    errorContext = null;
  }

  function handleClearCache() {
    if (typeof window !== "undefined") {
      try {
        // Access cache manager via window if exposed
        const cacheManager = (window as any).__RESOURCE_GRAPH_CACHE;
        if (cacheManager && typeof cacheManager.clearAll === "function") {
          cacheManager.clearAll();
        }

        // Also try localStorage directly
        const cacheKey = "rill.resourceGraph.v2";
        localStorage.removeItem(cacheKey);

        // Reset error and reload
        handleReset();
        window.location.reload();
      } catch (e) {
        console.error("Failed to clear cache:", e);
      }
    }
  }
</script>

{#if error && errorContext}
  {#if fallback}
    <svelte:component this={fallback} {error} />
  {:else}
    <div class="error-boundary">
      <div class="error-content">
        <div class="error-icon">
          {#if errorContext.severity === ErrorSeverity.CRITICAL}
            <span class="icon-critical">⚠️</span>
          {:else if errorContext.severity === ErrorSeverity.ERROR}
            <span class="icon-error">❌</span>
          {:else}
            <span class="icon-warning">⚠️</span>
          {/if}
        </div>

        <h3 class="error-title">
          {errorContext.severity === ErrorSeverity.CRITICAL
            ? "Critical Error"
            : "Something went wrong"}
        </h3>

        <p class="error-message">
          {getUserErrorMessage(error)}
        </p>

        {#if getRecoveryAction(error)}
          <p class="error-recovery">
            {getRecoveryAction(error)}
          </p>
        {/if}

        <div class="error-actions">
          <button class="btn-primary" on:click={handleReset}>
            Try Again
          </button>

          <button class="btn-secondary" on:click={handleClearCache}>
            Clear Cache
          </button>

          <button
            class="btn-secondary"
            on:click={() => window.location.reload()}
          >
            Reload Page
          </button>
        </div>

        {#if errorContext.data}
          <details class="error-details">
            <summary>Technical Details</summary>
            <pre>{JSON.stringify(
                {
                  component: errorContext.component,
                  operation: errorContext.operation,
                  message: error.message,
                  stack: error.stack,
                  data: errorContext.data,
                },
                null,
                2,
              )}</pre>
          </details>
        {/if}
      </div>
    </div>
  {/if}
{:else}
  <slot />
{/if}

<style lang="postcss">
  .error-boundary {
    @apply flex min-h-[400px] w-full items-center justify-center rounded-lg border border-red-200 bg-red-50 p-6;
  }

  .error-content {
    @apply max-w-md text-center;
  }

  .error-icon {
    @apply mb-4 text-5xl;
  }

  .icon-critical {
    @apply text-red-600;
  }

  .icon-error {
    @apply text-red-500;
  }

  .icon-warning {
    @apply text-orange-500;
  }

  .error-title {
    @apply mb-2 text-xl font-semibold text-red-900;
  }

  .error-message {
    @apply mb-2 text-base text-red-800;
  }

  .error-recovery {
    @apply mb-4 text-sm text-red-700;
  }

  .error-actions {
    @apply flex flex-wrap justify-center gap-2;
  }

  .btn-primary {
    @apply rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-fg-primary;
  }

  .btn-primary:hover {
    @apply bg-red-700;
  }

  .btn-primary:focus {
    @apply outline-none ring-2 ring-red-500 ring-offset-2;
  }

  .btn-secondary {
    @apply rounded-md border border-red-300 bg-surface-background px-4 py-2 text-sm font-medium text-red-700;
  }

  .btn-secondary:hover {
    @apply bg-red-50;
  }

  .btn-secondary:focus {
    @apply outline-none ring-2 ring-red-500 ring-offset-2;
  }

  .error-details {
    @apply mt-4 text-left;
  }

  .error-details summary {
    @apply cursor-pointer text-sm font-medium text-red-700;
  }

  .error-details summary:hover {
    @apply text-red-900;
  }

  .error-details pre {
    @apply mt-2 overflow-auto rounded bg-red-100 p-3 text-xs text-red-900;
    max-height: 200px;
  }
</style>
