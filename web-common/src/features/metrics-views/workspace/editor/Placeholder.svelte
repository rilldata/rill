<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
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
    V1GetCatalogEntryResponse,
    getRuntimeServiceGetCatalogEntryQueryKey,
    runtimeServiceGetCatalogEntry,
    runtimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let metricsName: string;
  export let view: EditorView = undefined;

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
      const model = await queryClient.fetchQuery<V1GetCatalogEntryResponse>({
        queryKey: getRuntimeServiceGetCatalogEntryQueryKey(
          $runtime?.instanceId,
          modelName
        ),
        queryFn: () =>
          runtimeServiceGetCatalogEntry($runtime?.instanceId, modelName),
      });

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
    /**
     * go ahead and optimistically update the editor view.
     */
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: str,
      },
      // tell the editor that this is a transaction that should _not_ be
      // debounced. This tells the binder to delay dispatching out of the editor component
      // any reconciliation update.
      annotations: skipDebounceAnnotation.of(true),
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

    /** optimistically update the editor. We will dispatch
     * a debounce annotation here to tell the MetricsWorkspace
     * not to debounce this update.
     */
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: yaml,
      },
      annotations: skipDebounceAnnotation.of(true),
    });

    /** invalidate and show results */
    invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
  }
</script>

<div class="whitespace-normal">
  Auto-generate a <WithTogglableFloatingElement
    inline
    let:toggleFloatingElement
    distance={8}
  >
    <button
      disabled={!$models?.data?.length}
      class={buttonClasses}
      on:click={toggleFloatingElement}
      >metrics configuration from an existing model</button
    >,
    <Menu
      dark
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
