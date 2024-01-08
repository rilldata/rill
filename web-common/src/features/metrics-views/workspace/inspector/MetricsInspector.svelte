<script lang="ts">
  import { _ } from "svelte-i18n";
  import ColumnProfile from "@rilldata/web-common/components/column-profile/ColumnProfile.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    useModel,
    useModels,
  } from "@rilldata/web-common/features/models/selectors";
  import ModelInspectorHeader from "@rilldata/web-common/features/models/workspace/inspector/ModelInspectorHeader.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    V1Resource,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import { getModelOutOfPossiblyMalformedYAML } from "../../utils";

  export let metricsDefName: string;

  let showColumns = true;
  let showModelInformation = true;

  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricsDefName, EntityType.MetricsDefinition)
  );
  $: yaml = $fileQuery.data?.blob || "";

  // get file.
  $: modelName = getModelOutOfPossiblyMalformedYAML(yaml)?.replace(/"/g, "");

  // check to see if this model name exists.
  $: modelQuery = useModel($runtime.instanceId, modelName);

  $: allModels = useModels($runtime.instanceId);

  let isValidModel = false;
  $: if ($allModels?.data?.entries) {
    isValidModel = $allModels?.data.some(
      (model) => model?.meta?.name?.name === modelName
    );
  }

  let entry: V1Resource;

  // refresh entry value only if the data has changed
  $: entry = $modelQuery?.data || entry;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
</script>

<div>
  {#if modelName && !$modelQuery?.isError && isValidModel}
    {#key modelName}
      <div class="pt-1 pb-2" use:listenToNodeResize>
        <div class="pl-4 pr-4">
          <CollapsibleSectionTitle
            tooltipText="model summary"
            bind:active={showModelInformation}
          >
            {$_('model-summary')}
          </CollapsibleSectionTitle>
        </div>
        {#if showModelInformation}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            <ModelInspectorHeader
              {modelName}
              containerWidth={$observedNode?.clientWidth}
            />
            <hr class:opacity-0={!showColumns} class="transition-opacity" />
          </div>
        {/if}
      </div>
    {/key}
    <div class="model-profile pb-4 pt-2">
      {#if entry && entry?.model?.spec?.sql?.trim()?.length}
        <div class="pl-4 pr-4">
          <CollapsibleSectionTitle
            tooltipText="selected columns"
            bind:active={showColumns}
          >
            {$_('model-columns')}
          </CollapsibleSectionTitle>
        </div>

        {#if showColumns && entry?.meta?.name?.name}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            <ColumnProfile
              objectName={entry?.meta?.name?.name}
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
        <p>{$_('lets-get-started')}</p>
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
