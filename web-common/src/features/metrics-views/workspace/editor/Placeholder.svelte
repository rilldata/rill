<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    generateDashboardYAMLForTable,
    initBlankDashboardYAML,
  } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import {
    V1GetResourceResponse,
    connectorServiceOLAPGetTable,
    getConnectorServiceOLAPGetTableQueryKey,
    getRuntimeServiceGetResourceQueryKey,
    runtimeServiceGetResource,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let metricsName: string;
  export let view: EditorView | undefined = undefined;

  $: models = useModelFileNames($runtime.instanceId);

  const queryClient = useQueryClient();

  const buttonClasses =
    "inline hover:font-semibold underline underline-offset-2";

  // FIXME: shouldn't these be generalized and used everywhere?
  async function onAutogenerateConfigFromModel(modelName: string) {
    const instanceId = $runtime?.instanceId;
    const model = await queryClient.fetchQuery<V1GetResourceResponse>({
      queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
        "name.name": modelName,
        "name.kind": ResourceKind.Model,
      }),
      queryFn: () =>
        runtimeServiceGetResource(instanceId, {
          "name.name": modelName,
          "name.kind": ResourceKind.Model,
        }),
    });
    const schemaResp = await queryClient.fetchQuery({
      queryKey: getConnectorServiceOLAPGetTableQueryKey({
        instanceId,
        table: model?.resource?.model?.state?.table,
        connector: model?.resource?.model?.state?.connector,
      }),
      queryFn: () =>
        connectorServiceOLAPGetTable({
          instanceId,
          table: model?.resource?.model?.state?.table,
          connector: model?.resource?.model?.state?.connector,
        }),
    });

    const isModel = true;
    const dashboardYAML = schemaResp?.schema
      ? generateDashboardYAMLForTable(modelName, isModel, schemaResp?.schema)
      : "";

    await runtimeServicePutFile(
      $runtime.instanceId,
      getFileAPIPathFromNameAndType(metricsName, EntityType.MetricsDefinition),
      {
        blob: dashboardYAML,
        create: true,
        createOnly: true,
      },
    );
    await waitForResource(
      queryClient,
      $runtime.instanceId,
      getFilePathFromNameAndType(metricsName, EntityType.MetricsDefinition),
    );
    /**
     * go ahead and optimistically update the editor view.
     */
    view?.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: dashboardYAML,
      },
      // tell the editor that this is a transaction that should _not_ be
      // debounced. This tells the binder to delay dispatching out of the editor component
      // any reconciliation update.
      annotations: skipDebounceAnnotation.of(true),
    });
  }

  // FIXME: shouldn't these be generalized and used everywhere?
  async function onCreateSkeletonMetricsConfig() {
    const yaml = initBlankDashboardYAML(metricsName);

    await runtimeServicePutFile(
      $runtime.instanceId,
      getFileAPIPathFromNameAndType(metricsName, EntityType.MetricsDefinition),
      {
        blob: yaml,
        create: true,
        createOnly: true,
      },
    );

    /** optimistically update the editor. We will dispatch
     * a debounce annotation here to tell the MetricsWorkspace
     * not to debounce this update.
     */
    view?.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: yaml,
      },
      annotations: skipDebounceAnnotation.of(true),
    });
  }
</script>

<div class="whitespace-normal">
  Auto-generate a <WithTogglableFloatingElement
    distance={8}
    inline
    let:toggleFloatingElement
  >
    <button
      class={buttonClasses}
      disabled={!$models?.data?.length}
      on:click={toggleFloatingElement}
      >metrics configuration from an existing model</button
    >,
    <Menu
      dark
      on:click-outside={toggleFloatingElement}
      on:escape={toggleFloatingElement}
      slot="floating-element"
      let:toggleFloatingElement
    >
      {#each $models?.data ?? [] as model}
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
    class={buttonClasses}
    on:click={async () => {
      onCreateSkeletonMetricsConfig();
    }}>start with a skeleton</button
  >, or just start typing.
</div>
