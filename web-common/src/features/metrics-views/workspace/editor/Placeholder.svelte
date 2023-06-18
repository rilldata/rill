<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    addQuickMetricsToDashboardYAML,
    initBlankDashboardYAML,
  } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
  import { useModelNames } from "@rilldata/web-common/features/models/selectors";
  import {
    runtimeServiceGetCatalogEntry,
    runtimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let metricsName: string;

  $: models = useModelNames($runtime.instanceId);

  const queryClient = useQueryClient();

  const buttonClasses =
    "inline hover:font-semibold underline underline-offset-2";

  async function onAutogenerateConfigFromModel(
    modelName: string,
    str = undefined
  ) {
    if (str === undefined) {
      const model = await runtimeServiceGetCatalogEntry(
        $runtime?.instanceId,
        modelName
      );
      str = addQuickMetricsToDashboardYAML("", model?.entry?.model);
    }

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: $runtime.instanceId,
      path: getFilePathFromNameAndType(
        metricsName,
        EntityType.MetricsDefinition
      ),
      blob: str,
      create: true,
      createOnly: true,
      strict: false,
    });

    /** invalidate and show */
    invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
  }

  async function onCreateSkeletonMetricsConfig() {
    const yaml = initBlankDashboardYAML(metricsName);

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: $runtime.instanceId,
      path: getFilePathFromNameAndType(
        metricsName,
        EntityType.MetricsDefinition
      ),
      blob: yaml,
      create: true,
      createOnly: true,
      strict: false,
    });

    /** invalidate and show */
    invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
  }
</script>

<!-- completely empty case -->
<div class="whitespace-normal">
  Auto-generate a <WithTogglableFloatingElement
    inline
    let:toggleFloatingElement
    let:active
  >
    <Tooltip distance={8} suppress={active}>
      <button
        disabled={!$models?.data?.length}
        class={buttonClasses}
        on:click={toggleFloatingElement}
        >metrics configuration from an existing model</button
      >,
      <TooltipContent slot="tooltip-content"
        >Select a data model and auto-generate the config</TooltipContent
      ></Tooltip
    >
    <Menu
      slot="floating-element"
      on:click-outside={toggleFloatingElement}
      on:escape={toggleFloatingElement}
    >
      {#each $models?.data as model}
        <MenuItem
          on:select={() => {
            onAutogenerateConfigFromModel(model);
            toggleFloatingElement();
          }}
        >
          {model}
        </MenuItem>
      {/each}
    </Menu>
  </WithTogglableFloatingElement>
  <button
    on:click={async () => {
      onCreateSkeletonMetricsConfig();
    }}
    class={buttonClasses}>start with a skeleton</button
  >, or just start typing.
</div>
