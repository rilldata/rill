<script lang="ts">
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import ModelInspectorHeader from "./ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;
  export let resourceIsReconciling: boolean;
  export let hasErrors: boolean;
  export let modelIsEmpty: boolean;

  let containerWidth: number;
</script>

{#if !modelIsEmpty}
  {#if resourceIsReconciling}
    <div class="mt-6">
      <ReconcilingSpinner />
    </div>
  {:else if !hasErrors}
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
