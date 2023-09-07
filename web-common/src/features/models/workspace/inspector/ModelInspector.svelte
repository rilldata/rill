<script lang="ts">
  import ColumnProfileProvider from "@rilldata/web-common/components/column-profile/ColumnProfileProvider.svelte";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    createRuntimeServiceGetCatalogEntry,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: emptyModel = useModelFileIsEmpty($runtime?.instanceId, modelName);

  $: getModel = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );
  let model: V1Model;
  $: model = $getModel?.data?.entry?.model;
</script>

{#if !$emptyModel?.data}
  <div>
    {#key modelName}
      <div use:listenToNodeResize>
        <ColumnProfileProvider objectName={modelName} sql={model?.sql}>
          <ModelInspectorHeader
            {modelName}
            containerWidth={$observedNode?.clientWidth}
          />
          <hr />
          <ModelInspectorModelProfile {modelName} />
        </ColumnProfileProvider>
      </div>
    {/key}
  </div>
{:else}
  <div class="px-4 py-24 italic ui-copy-disabled text-center">
    Model is empty.
  </div>
{/if}
