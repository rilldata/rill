<script lang="ts">
  import WarningIcon from "web-common/src/components/icons/WarningIcon.svelte";
  import type { MissingRequiredFilter } from "@rilldata/web-common/features/dashboards/filters/required/required-filters.ts";

  let {
    missingRequiredFilters,
  }: { missingRequiredFilters: MissingRequiredFilter[] } = $props();
</script>

<div class="w-full flex justify-center px-6 pt-24 pb-12">
  <div
    class="flex flex-col items-center text-center gap-y-3 px-8 py-10 rounded-lg border border-gray-200 bg-surface-subtle shadow-sm w-full max-w-lg"
    role="alert"
  >
    <WarningIcon size="32px" className="text-amber-500" />
    <h2 class="text-lg font-semibold text-fg-primary">
      Select a value to continue
    </h2>
    <p class="text-sm text-fg-secondary">
      This dashboard requires values for the following filter{missingRequiredFilters.length >
      1
        ? "s"
        : ""}:
    </p>
    <ul
      class="text-sm text-fg-primary flex flex-wrap justify-center gap-x-2 gap-y-1"
    >
      {#each missingRequiredFilters as missing (missing.key)}
        <li
          class="px-2 py-0.5 rounded-md bg-red-50 border border-red-200 text-red-700"
        >
          {missing.label}
        </li>
      {/each}
    </ul>
  </div>
</div>
