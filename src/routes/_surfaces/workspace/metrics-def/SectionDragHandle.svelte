<script lang="ts">
  import { getContext } from "svelte";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import { drag } from "$lib/drag";
  import {
    modelPreviewVisibilityTween,
    // modelPreviewVisible,
    layout,
    assetVisibilityTween,
    inspectorVisibilityTween,
    SIDE_PAD,
  } from "$lib/application-state-stores/layout-store";

  // import PreviewTable from "$lib/components/table/PreviewTable.svelte";
  // import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
  // import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
  // import type {
  //   DerivedModelStore,
  //   PersistentModelStore,
  // } from "$lib/application-state-stores/model-stores";
  // import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Portal from "$lib/components/Portal.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
  // const queryHighlight = getContext("rill:app:query-highlight");
  // const persistentModelStore = getContext(
  //   "rill:app:persistent-model-store"
  // ) as PersistentModelStore;
  // const derivedModelStore = getContext(
  //   "rill:app:derived-model-store"
  // ) as DerivedModelStore;

  // let showPreview = true;

  // let currentModel: PersistentModelEntity;
  // $: currentModel =
  //   $store?.activeEntity && $persistentModelStore?.entities
  //     ? $persistentModelStore.entities.find(
  //         (q) => q.id === $store.activeEntity.id
  //       )
  //     : undefined;
  // let currentDerivedModel: DerivedModelEntity;
  // $: currentDerivedModel =
  //   $store?.activeEntity && $derivedModelStore?.entities
  //     ? $derivedModelStore.entities.find((q) => q.id === $store.activeEntity.id)
  //     : undefined;

  // track innerHeight to calculate the size of the editor element.
  let innerHeight;
</script>

<Portal target=".body">
  <div
    class="fixed z-50 drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center ml-2 mr-2"
    style:bottom="{(1 - $modelPreviewVisibilityTween) *
      $layout.modelPreviewHeight}px"
    style:left="{(1 - $assetVisibilityTween) * $layout.assetsWidth + 16}px"
    style:right="{(1 - $inspectorVisibilityTween) * $layout.inspectorWidth +
      16}px"
    style:padding-left="{$assetVisibilityTween * SIDE_PAD}px"
    style:padding-right="{$inspectorVisibilityTween * SIDE_PAD}px"
    use:drag={{
      minSize: 200,
      maxSize: innerHeight - 200,
      side: "modelPreviewHeight",
      orientation: "vertical",
      reverse: true,
    }}
  >
    <div class="border-t border-gray-300" />
    <div class="absolute right-1/2 left-1/2 top-1/2 bottom-1/2">
      <div
        class="border-gray-400 border bg-white rounded h-1 w-8 absolute -translate-y-1/2"
      />
    </div>
  </div>
</Portal>
