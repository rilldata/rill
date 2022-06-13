<script lang="ts">
  import { getContext } from "svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
</script>

{#if useModelWorkspace}
  <ModelWorkspaceHeader />
  <ModelView />
{:else}
  <!-- FIXME: this placeholder is here to show where you would plug in another kind of workspace component -->
  <ModelWorkspaceHeader />
  <ModelView />
{/if}
