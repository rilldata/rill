<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { initBlankDashboardYAML } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
  import { useModels } from "@rilldata/web-common/features/models/selectors";
  import {
    type V1Resource,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "../../connectors/selectors";
  import { createDashboardFromTableInMetricsEditor } from "../ai-generation/generateMetricsView";

  export let metricsName: string;
  export let filePath: string;
  export let view: EditorView | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(instanceId);
  $: isModelingSupported = $isModelingSupportedForDefaultOlapDriver.data;
  $: models = useModels(instanceId);

  const buttonClasses =
    "inline hover:font-semibold underline underline-offset-2";

  async function onAutogenerateConfigFromModel(modelRes: V1Resource) {
    await createDashboardFromTableInMetricsEditor(
      instanceId,
      modelRes?.model?.state?.resultTable ?? "",
      filePath,
    );
  }

  // FIXME: shouldn't these be generalized and used everywhere?
  async function onCreateSkeletonMetricsConfig() {
    const yaml = initBlankDashboardYAML(metricsName);

    await runtimeServicePutFile(instanceId, {
      path: filePath,
      blob: yaml,
      create: true,
      createOnly: true,
    });

    /** optimistically update the editor. We will dispatch
     * a debounce annotation here to tell the MetricsWorkspace
     * not to debounce this update.
     */
    queueMicrotask(() => {
      view?.dispatch({
        changes: {
          from: 0,
          to: view.state.doc.length,
          insert: yaml,
        },
      });
    });
  }
</script>

<div class="whitespace-normal">
  {#if isModelingSupported}
    Auto-generate a
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild let:builder>
        <button
          use:builder.action
          {...builder}
          class={buttonClasses}
          disabled={!$models?.data?.length}
        >
          metrics configuration from an existing model
        </button>
      </DropdownMenu.Trigger>,
      <DropdownMenu.Content align="start" sameWidth>
        {#each $models?.data ?? [] as model, i (i)}
          {#if model?.model?.state?.resultTable}
            <DropdownMenu.Item
              on:click={() => {
                void onAutogenerateConfigFromModel(model);
              }}
            >
              {model?.model?.state?.resultTable}
            </DropdownMenu.Item>
          {/if}
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}

  <button
    class={buttonClasses}
    on:click={async () => {
      onCreateSkeletonMetricsConfig();
    }}
    >{#if isModelingSupported}s{:else}S{/if}tart with a skeleton</button
  >, or just start typing.
</div>
