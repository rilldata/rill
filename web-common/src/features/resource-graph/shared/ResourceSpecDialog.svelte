<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { getResourceSpec } from "./resource-actions";

  export let open = false;
  export let resourceName: string;
  export let kind: string | undefined;
  export let resource: V1Resource | undefined;

  $: specContent = getResourceSpec(resource);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>
        {resourceName}
        <span class="text-fg-tertiary font-normal text-sm ml-2"
          >{kind ? displayResourceKind(kind) : ""}</span
        >
      </Dialog.Title>
    </Dialog.Header>
    <div class="spec-container">
      {#if !resource}
        <p class="text-sm text-fg-secondary">No resource data available</p>
      {:else}
        <pre class="spec-content">{specContent}</pre>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .spec-container {
    @apply overflow-auto flex-1 min-h-0;
  }
  .spec-content {
    @apply text-xs font-mono whitespace-pre-wrap bg-surface-subtle rounded-md p-4;
  }
</style>
