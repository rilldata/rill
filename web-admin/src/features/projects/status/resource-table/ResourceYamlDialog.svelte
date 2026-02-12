<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { prettyResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let open = false;
  export let resourceName = "";
  export let resourceKind = "";
  export let resource: V1Resource | undefined = undefined;

  function getResourceSpec(res: V1Resource | undefined): string {
    if (!res) return "";

    // Get the typed spec based on the resource kind key
    const kindKeys = [
      "source",
      "model",
      "metricsView",
      "explore",
      "theme",
      "component",
      "canvas",
      "api",
      "connector",
      "report",
      "alert",
    ] as const;

    for (const key of kindKeys) {
      if (res[key]) {
        return JSON.stringify(res[key], null, 2);
      }
    }

    // Fallback: show the full resource minus meta
    const { meta: _meta, ...rest } = res;
    return JSON.stringify(rest, null, 2);
  }

  $: specContent = getResourceSpec(resource);
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>
        {resourceName}
        <span class="text-fg-tertiary font-normal text-sm ml-2"
          >{prettyResourceKind(resourceKind)}</span
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
