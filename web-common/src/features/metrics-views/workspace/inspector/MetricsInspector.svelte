<script lang="ts">
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
    createRuntimeServiceListCatalogEntries,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import { getModelOutOfPossiblyMalformedYAML } from "../../utils";

  export let metricsDefName: string;

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );
  $: yaml = $fileQuery.data?.blob || "";

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");

  // check to see if this model name exists.
  $: modelQuery = createRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    modelName
  );

  $: allModels = createRuntimeServiceListCatalogEntries($runtime.instanceId, {
    type: "OBJECT_TYPE_MODEL",
  });

  let isValidModel = false;
  $: if ($allModels?.data?.entries) {
    isValidModel = $allModels?.data.entries.some(
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
    <div class="model-profile pb-4 pt-2">
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
            <ColumnProfile
              key={entry?.model?.sql}
              objectName={entry?.model?.name}
              indentLevel={0}
            />
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
