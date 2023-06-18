<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
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

  // FIXME: shouldn't these be generalized and used everywhere?
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

    /** invalidate and show results */
    invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
  }

  // FIXME: shouldn't these be generalized and used everywhere?
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

    /** invalidate and show results */
    invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
  }
</script>

<div class="whitespace-normal">
  Auto-generate a <WithTogglableFloatingElement
    inline
    let:toggleFloatingElement
  >
    <button
      disabled={!$models?.data?.length}
      class={buttonClasses}
      on:click={toggleFloatingElement}
      >metrics configuration from an existing model</button
    >,
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
