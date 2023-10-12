<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModel, useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  $: modelQuery = useModel($runtime?.instanceId, modelName);

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: emptyModel = useModelFileIsEmpty($runtime?.instanceId, modelName);
</script>

{#if !$emptyModel?.data}
  {#if resourceIsLoading($modelQuery?.data)}
    <div class="mt-6">
      <ReconcilingSpinner />
    </div>
  {:else}
    <div>
      {#key modelName}
        <div use:listenToNodeResize>
          <ModelInspectorHeader
            {modelName}
            containerWidth={$observedNode?.clientWidth}
          />
          <hr />
          <ModelInspectorModelProfile {modelName} />
        </div>
      {/key}
    </div>
  {/if}
{:else}
  <div class="px-4 py-24 italic ui-copy-disabled text-center">
    Model is empty.
  </div>
{/if}
