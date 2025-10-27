<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";

  export let resource: V1Resource;

  const name = resource?.meta?.name?.name ?? "unknown";
  const kind = resource?.meta?.name?.kind ?? "";
  const error = resource?.meta?.reconcileError ?? "";

  function close() {
    overlay.clear();
  }
</script>

<div class="dialog">
  <div class="header">
    <h3 class="title">{kind ? kind.replace(/^rill\.runtime\.v1\./, "") : "Resource"} “{name}” error</h3>
    <button class="close" on:click={close} aria-label="Close">✕</button>
  </div>
  {#if error}
    <pre class="error" data-testid="resource-error">{error}</pre>
  {:else}
    <p class="no-error">No error message available.</p>
  {/if}
  <div class="footer">
    <button class="btn" on:click={close}>Close</button>
  </div>
  
</div>

<style lang="postcss">
  .dialog {
    @apply w-[720px] max-w-full rounded-lg border border-gray-200 bg-white shadow-lg flex flex-col;
  }
  .header {
    @apply flex items-center justify-between px-4 py-3 border-b border-gray-200;
  }
  .title {
    @apply text-sm font-semibold text-foreground;
  }
  .close {
    @apply h-7 w-7 rounded border border-gray-300 bg-white text-sm text-gray-600 hover:bg-gray-50 hover:text-gray-800;
    line-height: 1.25rem;
  }
  .error {
    @apply m-4 p-3 rounded bg-gray-50 text-sm text-gray-800 overflow-auto whitespace-pre-wrap border border-gray-200;
    max-height: 50vh;
  }
  .no-error {
    @apply m-4 text-sm text-gray-600;
  }
  .footer {
    @apply flex justify-end gap-2 px-4 py-3 border-t border-gray-200;
  }
  .btn {
    @apply h-8 px-3 rounded border border-gray-300 bg-white text-sm text-gray-700 hover:bg-gray-50;
  }
</style>

