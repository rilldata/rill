<script lang="ts">
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import type {
    V1ParseError,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import WorkspaceEditorContainer from "./WorkspaceEditorContainer.svelte";

  export let resource: V1Resource | undefined = undefined;
  export let parseError: V1ParseError | undefined = undefined;
  export let remoteContent: string | null | undefined = undefined;

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

  $: error = parseError?.message ?? rootCauseReconcileError;

  interface $$Slots {
    default: { rootCauseReconcileError: string | undefined };
  }
</script>

<WorkspaceEditorContainer {error} showError={!!remoteContent}>
  <slot {rootCauseReconcileError} />
</WorkspaceEditorContainer>
