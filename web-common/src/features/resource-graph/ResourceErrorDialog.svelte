<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { prettyResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;
  export let open = true;

  const name = resource?.meta?.name?.name ?? "unknown";
  const kind = prettyResourceKind(resource?.meta?.name?.kind ?? "");
  const error = resource?.meta?.reconcileError ?? "";
</script>

<AlertDialog bind:open>
  <AlertDialogContent class="max-w-[720px]">
    <AlertDialogHeader>
      <AlertDialogTitle>
        {kind || "Resource"} "{name}" error
      </AlertDialogTitle>
    </AlertDialogHeader>
    <AlertDialogDescription asChild>
      <div>
        {#if error}
          <pre
            class="m-0 max-h-[50vh] overflow-auto whitespace-pre-wrap rounded border border-gray-200 bg-gray-50 p-3 text-sm text-gray-800"
            data-testid="resource-error">{error}</pre>
        {:else}
          <p class="text-sm text-gray-600">No error message available.</p>
        {/if}
      </div>
    </AlertDialogDescription>
    <AlertDialogFooter>
      <Button on:click={() => (open = false)}>Close</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
