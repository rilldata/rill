<script lang="ts">
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import { createResizeListenerActionFactory } from "$lib/components/actions/create-resize-listener-factory";
  import StickToHeaderDivider from "$lib/components/panel/StickToHeaderDivider.svelte";
  import { getContext } from "svelte";
  import ModelInspectorHeader from "./header/ModelInspectorHeader.svelte";
  import ModelInspectorModelProfile from "./ModelInspectorModelProfile.svelte";

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  const store = getContext("rill:app:store") as ApplicationStore;

  /** Select the explicit ID to prevent unneeded reactive updates in currentModel */
  $: activeEntityID = $store?.activeEntity?.id;

  let currentModel: PersistentModelEntity;
  $: currentModel =
    activeEntityID && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find((q) => q.id === activeEntityID)
      : undefined;

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
</script>

{#key currentModel?.id}
  <div use:listenToNodeResize>
    <ModelInspectorHeader containerWidth={$observedNode?.clientWidth} />
    <StickToHeaderDivider />
    <ModelInspectorModelProfile />
  </div>
{/key}
