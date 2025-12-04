<script lang="ts">
  import {
    handleCanvasStoreInitialization, // Your async factory
  } from "../state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { page } from "$app/stores";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";

  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: ({ url } = $page);

  $: initPromise = handleCanvasStoreInitialization(canvasName, instanceId);
</script>

{#await initPromise}
  <Spinner status={EntityStatus.Running} size="32px" />
{:then resolvedStore}
  {@const store = resolvedStore.store}
  {@const _ = store.canvasEntity
    .onUrlChange({ url, loadFunction: false })
    .catch(console.error)}
  <slot {_} />
{:catch error}
  <ErrorPage
    header="Error loading canvas"
    body={error.message}
    statusCode={500}
  />
{/await}
