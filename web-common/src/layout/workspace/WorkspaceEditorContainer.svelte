<script lang="ts">
  import { slide } from "svelte/transition";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import type {
    V1ParseError,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExplainAndFixErrorButton from "@rilldata/web-common/features/chat/ExplainAndFixErrorButton.svelte";

  // Direct error string (existing API, still supported)
  export let error: string | undefined = undefined;
  export let showError = true;

  // Resource-based error derivation (replaces WorkspaceError wrapper)
  export let resource: V1Resource | undefined = undefined;
  export let parseError: V1ParseError | undefined = undefined;
  export let remoteContent: string | null | undefined = undefined;
  export let filePath: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();

  $: reconcileError = resource?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    resource,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;

  $: derivedError = parseError?.message ?? rootCauseReconcileError;
  $: effectiveError = error ?? derivedError;
  $: effectiveShowError =
    remoteContent !== undefined ? !!remoteContent : showError;
</script>

<div
  class="flex flex-col size-full gap-y-1 bg-surface-subtle rounded-[2px] border overflow-hidden"
>
  <div
    class="size-full relative overflow-hidden flex flex-col items-center justify-center"
    class:!border-red-500={effectiveError}
  >
    <slot />
  </div>

  {#if effectiveError && effectiveShowError}
    <div
      role="status"
      transition:slide={{ duration: LIST_SLIDE_DURATION }}
      class="border border-destructive bg-destructive/15 dark:bg-destructive/30 text-fg-primary border-l-4 px-3 py-2.5 text-sm flex-shrink-0"
    >
      <div class="flex gap-x-2">
        <CancelCircle className="text-destructive flex-shrink-0 mt-0.5" />
        <div class="flex flex-col gap-2 min-w-0">
          <span class="break-words">{effectiveError}</span>
          {#if filePath}
            <ExplainAndFixErrorButton {filePath} />
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
