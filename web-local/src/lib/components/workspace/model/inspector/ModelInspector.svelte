<script lang="ts">
  import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { createResizeListenerActionFactory } from "@rilldata/web-local/lib/components/actions/create-resize-listener-factory";
  import StickToHeaderDivider from "@rilldata/web-local/lib/components/panel/StickToHeaderDivider.svelte";
  import { getContext } from "svelte";
  import ModelInspectorHeader from "./header/ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  export let modelName: string;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  let currentModel: PersistentModelEntity;
  $: currentModel = $persistentModelStore?.entities
    ? $persistentModelStore.entities.find((q) => q.name === modelName)
    : undefined;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
</script>

{#key currentModel?.id}
  <div use:listenToNodeResize>
    <ModelInspectorHeader
      {modelName}
      containerWidth={$observedNode?.clientWidth}
    />
    <StickToHeaderDivider />
    <ModelInspectorModelProfile />
  </div>
{/key}
