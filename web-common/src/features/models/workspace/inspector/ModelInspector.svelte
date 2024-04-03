<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { useModelFileIsEmpty } from "../../selectors";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let filePath: string;

  const queryClient = useQueryClient();

  let containerWidth: number;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: modelQuery = fileArtifact.getResource(queryClient, $runtime.instanceId);
  $: modelHasError = fileArtifact.getHasErrors(
    queryClient,
    $runtime.instanceId,
  );

  $: emptyModel = useModelFileIsEmpty($runtime?.instanceId, filePath);
</script>

{#if !$emptyModel?.data}
  {#if resourceIsLoading($modelQuery?.data)}
    <div class="mt-6">
      <ReconcilingSpinner />
    </div>
  {:else if !$modelHasError}
    <div>
      {#key filePath}
        <div bind:clientWidth={containerWidth}>
          <ModelInspectorHeader {filePath} {containerWidth} />
          <hr />
          <ModelInspectorModelProfile {filePath} />
        </div>
      {/key}
    </div>
  {/if}
{:else}
  <div class="px-4 py-24 italic ui-copy-disabled text-center">
    Model is empty.
  </div>
{/if}
