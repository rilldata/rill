<script lang="ts">
  import AlertCircle from "../../../../components/icons/AlertCircle.svelte";

  export let headline: string;
  export let error: string | null | undefined = undefined;
</script>

<div class="rounded-lg border border-gray-200 bg-gray-50 p-3 text-sm">
  <div class="flex items-start gap-2">
    <div class="mt-0.5 text-red-500">
      <AlertCircle size="14px" />
    </div>
    <div class="font-medium text-gray-900">{headline}</div>
  </div>
  {#if error}
    <details class="group mt-2">
      <summary
        class="flex cursor-pointer items-center gap-1.5 text-xs text-gray-600 hover:text-gray-900 hover:underline"
      >
        <svg
          class="h-3 w-3 transition-transform group-open:rotate-90"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 5l7 7-7 7"
          />
        </svg>
        <span>Show details</span>
      </summary>
      <div
        class="mt-2 rounded border border-gray-200 bg-white p-2 font-mono text-xs text-gray-700"
      >
        {error}
      </div>
    </details>
  {/if}
</div>
