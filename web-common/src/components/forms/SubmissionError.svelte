<script lang="ts">
  import { AlertCircleIcon } from "lucide-svelte";

  export let message: string;
  export let details: string | undefined = undefined;
</script>

<div class="error-container">
  <div class="flex items-start gap-1">
    <div class="flex-shrink-0 flex items-start">
      <AlertCircleIcon size={22} class="text-red-600" />
    </div>
    <div class="flex-1">
      <div class="message text-gray-700 font-normal text-sm">{message}</div>
      {#if details}
        <div class="details-section">
          <pre class="details">{details}</pre>
        </div>
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .error-container {
    @apply border-red-600;
    @apply p-2 flex border rounded mb-3;
  }

  .details-section {
    @apply mt-2 text-xs;
  }

  .details {
    @apply bg-gray-50 p-2 rounded text-gray-600 whitespace-pre-wrap break-words;
    @apply border border-gray-200;
  }
</style>
