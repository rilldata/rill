<script lang="ts">
  import AlertCircle from "../../../../components/icons/AlertCircle.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";

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
        <div class="transition-transform group-open:rotate-90">
          <ChevronRight size="12px" />
        </div>
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
