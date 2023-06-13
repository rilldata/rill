<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import { getModelOutOfPossiblyMalformedYAML } from "../../utils";

  export let yaml: string;
  export let metricsDefName: string;

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml).replace(/"/g, "");

  // check to see if this model name exists.
  //$: modelExists = $fileArtifactsStore.has(modelName);
  $: modelQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );

  $: models = createRuntimeServiceListCatalogEntries($runtime.instanceId, {
    type: "OBJECT_TYPE_MODEL",
  });

  let isValidModel = false;
  $: if ($models?.data?.entries) {
    isValidModel = $models?.data.entries.some(
      (model) => model.name === modelName
    );
  }

  $: getModel = createRuntimeServiceGetCatalogEntry(
    $runtime?.instanceId,
    modelName
  );
  let entry;
  // refresh entry value only if the data has changed
  $: entry = $getModel?.data?.entry || entry;
  let showColumns = true;
</script>

<div>
  {#if modelName && !$modelQuery?.isError && isValidModel}
    <div class="model-profile pb-4 pt-4">
      {#if entry && entry?.model?.sql?.trim()?.length}
        <div class="pl-4 pr-4">
          <CollapsibleSectionTitle
            tooltipText="selected columns"
            bind:active={showColumns}
          >
            Model columns
          </CollapsibleSectionTitle>
        </div>

        {#if showColumns}
          <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
            <!-- {#key entry?.model?.sql} -->
            <ColumnProfile
              key={entry?.model?.sql}
              objectName={entry?.model?.name}
              indentLevel={0}
            />
            <!-- {/key} -->
          </div>
        {/if}
      {/if}
    </div>
  {:else}
    <div
      class="px-4 py-24 italic ui-copy-disabled text-center"
      style:text-wrap="balance"
      style:max-inline-size="50ch"
    >
      {#if !yaml?.length}
        <p>Let's get started.</p>
        <p>add <code>model: MODEL_NAME</code> to connect a Model.</p>
      {:else if modelName !== undefined}
        <div>
          <p>Model not defined.</p>
          <p>
            Set a model with <code>model: MODEL_NAME</code> to connect your metrics
            to a model.
          </p>
        </div>
      {/if}
    </div>
  {/if}
</div>
