<script lang="ts">
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { PivotQueryError } from "./types";

  export let errors: PivotQueryError[];

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
</script>

<div class="flex flex-col items-center w-full h-full">
  <span class="text-3xl font-normal m-2">Sorry, unexpected query error!</span>
  {#if traceIds.length > 0}
    <div class="text-sm text-fg-primary mt-1">
      <b>Trace ID{traceIds.length !== 1 ? "s" : ""}</b>: {traceIds.join(", ")}
    </div>
  {/if}
  <div class="text-base text-fg-secondary mt-4">
    One or more APIs failed with the following error{uniqueErrors.length !== 1
      ? "s"
      : ""}:
  </div>

  {#each uniqueErrors as error}
    <div class="flex text-base gap-x-2 items-start max-w-4xl">
      <span class="text-red-600 font-semibold whitespace-nowrap"
        >{error.statusCode}:</span
      >
      <span class="text-fg-primary flex-1 break-words">
        {truncateMessage(error.message ?? "")}
      </span>
      <button
        class="flex-shrink-0 p-1 hover:bg-gray-100 rounded transition-colors cursor-pointer"
        on:click={() => handleCopyError(error)}
        title="Copy full error message"
        aria-label="Copy error message to clipboard"
      >
        <CopyIcon size="16px" color="#6B7280" />
      </button>
    </div>
  {/each}
</div>
