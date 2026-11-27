<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { PivotQueryError } from "./types";

  export let errors: PivotQueryError[];
  // The height of the space the error is shown in, from a host that limits it.
  export let availableHeight: number | undefined = undefined;

  const MAX_ERROR_LENGTH = 500;

  function removeDuplicates(errors: PivotQueryError[]): PivotQueryError[] {
    const seen = new Set();
    return errors.filter((error) => {
      const key = `${error.statusCode}-${error.message}`;
      if (seen.has(key)) {
        return false;
      } else {
        seen.add(key);
        return true;
      }
    });
  }

  function truncateMessage(message: string): string {
    if (message.length <= MAX_ERROR_LENGTH) {
      return message;
    }
    return message.slice(0, MAX_ERROR_LENGTH) + "...";
  }

  function handleCopyError(error: PivotQueryError) {
    const fullError = `${error.statusCode}${error.traceId ? ` (Trace ID: ${error.traceId})` : ""}: ${error.message}`;
    copyToClipboard(fullError, "Error message copied to clipboard");
  }

  function getUniqueTraceIds(errors: PivotQueryError[]): string[] {
    const traceIds = errors
      .map((error) => error.traceId)
      .filter((id): id is string => id !== undefined && id !== null);
    return [...new Set(traceIds)];
  }

  $: traceIds = getUniqueTraceIds(errors);

  let uniqueErrors = removeDuplicates(errors);

  // The height the heading, the summary line and one error need.
  // A shorter space, such as a canvas table sized for a few rows,
  // gets a small heading and no summary, so the error itself stays in view.
  const FULL_LAYOUT_HEIGHT = 120;

  $: compact =
    availableHeight !== undefined && availableHeight < FULL_LAYOUT_HEIGHT;
</script>

<div class="pivot-error" class:compact>
  <span class="heading">{m.pivot_unexpected_query_error()}</span>
  {#if traceIds.length > 0}
    <div class="text-sm text-fg-primary mt-1">
      <b>{m.pivot_trace_ids_label({ count: traceIds.length })}</b>: {traceIds.join(
        ", ",
      )}
    </div>
  {/if}
  <div class="summary">
    {m.pivot_apis_failed({ count: uniqueErrors.length })}
  </div>

  {#each uniqueErrors as error}
    <div class="error">
      <span class="text-red-600 font-semibold whitespace-nowrap"
        >{error.statusCode}:</span
      >
      <span class="text-fg-primary flex-1 break-words">
        {truncateMessage(error.message ?? "")}
      </span>
      <button
        class="flex-shrink-0 p-1 hover:bg-surface-hover rounded transition-colors cursor-pointer"
        onclick={() => handleCopyError(error)}
        title={m.dashboard_copy_error()}
        aria-label={m.dashboard_copy_error_to_clipboard_aria()}
      >
        <CopyIcon size="16px" color="#6B7280" />
      </button>
    </div>
  {/each}
</div>

<style lang="postcss">
  .pivot-error {
    @apply flex flex-col items-center w-full h-full overflow-y-auto;
  }

  .heading {
    @apply text-3xl font-normal m-2;
  }

  .summary {
    @apply text-base text-fg-secondary mt-4;
  }

  .error {
    @apply flex text-base gap-x-2 items-start max-w-4xl;
  }

  .compact .heading {
    @apply text-sm font-semibold m-1;
  }

  .compact .summary {
    @apply hidden;
  }

  .compact .error {
    @apply text-sm px-2;
  }
</style>
