<script lang="ts">
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PivotColumnTruncation } from "./types";

  export let columnTruncation: PivotColumnTruncation;

  $: ({ shown, total } = columnTruncation);
</script>

<div class="notice" role="status">
  <InfoCircle size="14px" className="flex-none" />
  <span>
    {total === undefined
      ? m.pivot_columns_truncated_unknown_total({ shown })
      : m.pivot_columns_truncated({ shown, total })}
  </span>
</div>

<style lang="postcss">
  .notice {
    @apply flex flex-none items-center gap-x-1.5 w-fit max-w-full;
    @apply rounded-sm border px-2 py-1 text-xs;
    @apply border-yellow-200 bg-yellow-50 text-yellow-800;
    @apply dark:border-yellow-900/40 dark:bg-yellow-900/20 dark:text-yellow-300;
  }
</style>
