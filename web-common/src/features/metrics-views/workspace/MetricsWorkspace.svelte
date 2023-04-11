<script lang="ts">
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    V1PutFileAndReconcileResponse,
    createRuntimeServicePutFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ConfigInspector from "./ConfigInspector.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";
  import YAMLEditor from "./YAMLEditor.svelte";
  import {
    createPlaceholderElement,
    rillEditorPlaceholder,
  } from "./rill-editor-placeholder";
  // the runtime yaml string
  export let yaml: string;
  export let metricsDefName: string;

  // this store is used to store errors that are not related to the reconciliation/runtime
  // used to prevent the user from going to the dashboard.
  // Ultimately, the runtime should be catching the different errors we encounter with regards to
  // mismatches between the fields. For now, this is a very simple to use solution.
  let configurationErrorStore = writable({
    defaultTimeRange: null,
    smallestTimeGrain: null,
    model: null,
    timeColumn: null,
  });
  setContext("rill:metrics-config:errors", configurationErrorStore);

  const queryClient = useQueryClient();
  const { listenToNodeResize } = createResizeListenerActionFactory();

  $: instanceId = $runtime.instanceId;

  const switchToMetrics = async (metricsDefName: string) => {
    if (!metricsDefName) return;

    appStore.setActiveEntity(metricsDefName, EntityType.MetricsDefinition);
  };

  $: switchToMetrics(metricsDefName);

  const metricMigrate = createRuntimeServicePutFileAndReconcile();
  async function callReconcileAndUpdateYaml(internalYamlString) {
    const filePath = getFilePathFromNameAndType(
      metricsDefName,
      EntityType.MetricsDefinition
    );
    const resp = (await $metricMigrate.mutateAsync({
      data: {
        instanceId,
        path: filePath,
        blob: yaml,
        create: false,
      },
    })) as V1PutFileAndReconcileResponse;
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    invalidateAfterReconcile(queryClient, $runtime.instanceId, resp);
  }

  function updateYAML(event) {
    const { content } = event.detail;
    callReconcileAndUpdateYaml(instanceId, metricsDefName, content);
  }

  const placeholderSet = createPlaceholderElement(yaml);
  const placeholder = rillEditorPlaceholder(placeholderSet.DOMElement);
  $: placeholderSet.set(yaml);
  //placeholderSet.
  placeholderSet.on("test", (event) => {
    console.log(event.detail);
  });
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} {yaml} />
  <div slot="body" use:listenToNodeResize>
    <div
      class="editor-pane bg-gray-100 p-6 flex flex-col"
      style:height="calc(100vh - var(--header-height))"
    >
      <div class="overflow-y-auto bg-white p-2 rounded">
        <YAMLEditor
          content={yaml}
          on:update={updateYAML}
          plugins={[placeholder]}
        />
      </div>
    </div>
  </div>
  <ConfigInspector slot="inspector" {metricsDefName} {yaml} />
</WorkspaceContainer>
