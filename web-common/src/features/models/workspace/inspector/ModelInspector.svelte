<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModel, useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const queryClient = useQueryClient();

  let containerWidth: number;

  $: filePath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: modelQuery = useModel($runtime?.instanceId, modelName);
  $: modelHasError = fileArtifact.getHasErrors(
    queryClient,
    $runtime?.instanceId,
  );

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
        <div bind:clientWidth={containerWidth}>
          <ModelInspectorHeader {modelName} {containerWidth} />
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
