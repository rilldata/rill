<!--
  This file checks for the existence of a `rill.yaml` file and handles the corresponding scenarios:
  - If the file exists, the app continues as normal.
  - If the file does not exist, the user is redirected to the Welcome page (DuckDB projects) or the project is initialized immediately (Clickhouse and Druid projects).

  We perform the check in two different ways:
  - onMount: Ensures that on the initial page load, we only proceed after the check is complete to avoid any content flash.
  - continuously: If the user deletes the `rill.yaml` file, they are immediately redirected to the Welcome page.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import {
    createRuntimeServiceUnpackEmpty,
    runtimeServiceGetInstance,
  } from "../../runtime-client";
  import { EMPTY_PROJECT_TITLE } from "./constants";
  import {
    isProjectInitialized,
    useIsProjectInitialized,
  } from "./is-project-initialized";

  // Initial check onMount
  let ready = false;
  onMount(async () => {
    const initialized = await isProjectInitialized($runtime.instanceId);

    if (initialized) {
      ready = true;
      return;
    }

    await handleUninitializedProject();
  });

  // Continuous check
  $: isProjectInitializedQuery = useIsProjectInitialized($runtime.instanceId);
  $: if ($isProjectInitializedQuery.data === false) {
    handleUninitializedProject();
  }

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleUninitializedProject() {
    // If the project is not initialized, determine what page to route to dependent on the OLAP connector
    const instance = await runtimeServiceGetInstance($runtime.instanceId);
    const olapConnector = instance.instance?.olapConnector;
    if (!olapConnector) {
      throw new Error("OLAP connector is not defined");
    }

    // DuckDB-backed projects should head to the Welcome page for user-guided initialization
    if (olapConnector === "duckdb") {
      ready = true;
      await goto("/welcome");
      return;
    }

    // Clickhouse and Druid-backed projects should be initialized immediately
    await $unpackEmptyProject.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        title: EMPTY_PROJECT_TITLE,
        force: true,
      },
    });
    ready = true;
  }
</script>

{#if ready}
  <slot />
{/if}
