<script lang="ts">
  import { _ } from "svelte-i18n";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModel, useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const queryClient = useQueryClient();

  $: path = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelQuery = useModel($runtime?.instanceId, modelName);
  $: modelHasError = getFileHasErrors(queryClient, $runtime?.instanceId, path);

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();

  $: emptyModel = useModelFileIsEmpty($runtime?.instanceId, modelName);
</script>

{#if !$emptyModel?.data}
  {#if resourceIsLoading($modelQuery?.data)}
    <div class="mt-6">
      <ReconcilingSpinner />
    </div>
  {:else if !$modelHasError}
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
    {$_('model-is-empty')}.
  </div>
{/if}
