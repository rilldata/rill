<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { skipDebounceAnnotation } from "@rilldata/web-common/components/editor/annotations";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { initBlankDashboardYAML } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../../../tables/selectors";
  import { createDashboardFromTableInMetricsEditor } from "../../ai-generation/generateMetricsView";

  export let metricsName: string;
  export let view: EditorView | undefined = undefined;

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: models = useModelFileNames($runtime.instanceId);

  const buttonClasses =
    "inline hover:font-semibold underline underline-offset-2";

  async function onAutogenerateConfigFromModel(modelName: string) {
    await createDashboardFromTableInMetricsEditor(
      $runtime.instanceId,
      modelName,
      metricsName,
    );
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
  {#if $isModelingSupportedForCurrentOlapDriver.data}
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
              void onAutogenerateConfigFromModel(model);
              toggleFloatingElement();
            }}
          >
            {model}
          </MenuItem>
        {/each}
      </Menu>
    </WithTogglableFloatingElement>
  {/if}
  <button
    class={buttonClasses}
    on:click={async () => {
      onCreateSkeletonMetricsConfig();
    }}
    >{#if $isModelingSupportedForCurrentOlapDriver.data}s{:else}S{/if}tart with
    a skeleton</button
  >, or just start typing.
</div>
