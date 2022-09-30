<script lang="ts">
  import type { PersistentModelStore } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import Model from "@rilldata/web-local/lib/components/workspace/Model.svelte";
  import { getContext } from "svelte";

  export let data;
  // check the modelId against the store.
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  $: console.log($persistentModelStore);

  $: if (
    $persistentModelStore &&
    !$persistentModelStore?.entities?.some((model) => model.id === data.modelId)
  ) {
    console.log("ok");
  }
</script>

<svelte:head>
  <!-- TODO: add the model name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<Model modelId={data.modelId} />
