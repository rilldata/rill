<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let error: string | undefined;
</script>

{#if error}
  <details class="group min-w-0 w-full whitespace-normal text-red-600">
    <summary
      class="flex cursor-pointer list-none items-center gap-x-1.5 rounded-sm font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring-focus [&::-webkit-details-marker]:hidden"
    >
      <CaretDownIcon
        size="14"
        className="shrink-0 -rotate-90 transition-transform group-open:rotate-0"
      />
      <span class="group-open:hidden">{m.error_show_details()}</span>
      <span class="hidden group-open:inline">{m.error_hide_details()}</span>
    </summary>
    <pre
      class="mt-2 max-h-48 overflow-auto whitespace-pre-wrap break-words [overflow-wrap:anywhere] text-xs font-mono">{error}</pre>
  </details>
{:else}
  <span>-</span>
{/if}
