<script lang="ts">
  import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import StickToHeaderDivider from "$lib/components/panel/StickToHeaderDivider.svelte";
  import { getContext, onMount } from "svelte";
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

  let containerWidth = 0;
  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container.clientWidth;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });
</script>

{#key currentModel?.id}
  <div bind:this={container}>
    <ModelInspectorHeader {containerWidth} />
    <StickToHeaderDivider />
    <ModelInspectorModelProfile />
  </div>
{/key}
