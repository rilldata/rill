<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import YamlPreview from "@rilldata/web-common/features/sources/modal/YamlPreview.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { stringify } from "yaml";

  export let open = false;
  export let resource: V1Resource | null = null;
  export let onClose: () => void = () => {};

  $: modelName = resource?.meta?.name?.name ?? "Model";

  // Build a clean object for YAML display, similar to `rill project describe`
  $: modelInfo = resource
    ? {
        name: resource.meta?.name?.name,
        kind: resource.meta?.name?.kind,
        filePaths: resource.meta?.filePaths,
        spec: resource.model?.spec,
        state: {
          executorConnector: resource.model?.state?.executorConnector,
          resultConnector: resource.model?.state?.resultConnector,
          resultTable: resource.model?.state?.resultTable,
          resultSchema: resource.model?.state?.resultSchema,
          partitionsModelId: resource.model?.state?.partitionsModelId,
          partitionsHaveErrors: resource.model?.state?.partitionsHaveErrors,
        },
      }
    : null;

  $: yamlContent = modelInfo
    ? stringify(modelInfo, { indent: 2, lineWidth: 0 })
    : "";
</script>

<Dialog.Root
  {open}
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
>
  <Dialog.Content class="max-w-2xl max-h-[80vh] overflow-y-auto">
    <Dialog.Header>
      <Dialog.Title>Model: {modelName}</Dialog.Title>
    </Dialog.Header>

    {#if yamlContent}
      <YamlPreview title="Model Information" yaml={yamlContent} />
    {:else}
      <div class="text-gray-500 py-4">No model information available</div>
    {/if}
  </Dialog.Content>
</Dialog.Root>
