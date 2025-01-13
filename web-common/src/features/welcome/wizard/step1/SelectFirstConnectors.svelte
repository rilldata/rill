<script lang="ts">
  import { createRuntimeServiceUnpackEmpty } from "../../../../runtime-client";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import type { OlapDriver } from "../../../connectors/olap/olap-config";
  import { EMPTY_PROJECT_TITLE } from "../../constants";
  import PickOlapManagement from "./A_PickOLAPManagement.svelte";
  import PickOlap from "./B_PickOLAP.svelte";
  import PickFirstSource from "./C_PickFirstSource.svelte";

  export let managementType: "self-managed" | "rill-managed";
  export let olapDriver: OlapDriver;
  export let firstDataSource: string | undefined;
  export let onSelectManagementType: (
    managementType: "self-managed" | "rill-managed",
  ) => void;
  export let onSelectOLAP: (olap: OlapDriver) => void;
  export let onSelectFirstDataSource: (dataSource: string) => void;
  export let onContinue: () => void;

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  function handleSkipFirstSource() {
    // unpack the empty project
    $unpackEmptyProject.mutate({
      instanceId: $runtime.instanceId,
      data: {
        displayName: EMPTY_PROJECT_TITLE,
      },
    });

    // create an OLAP connector file
    // Edit the rill.yaml file
  }
</script>

<h1 class="text-lead text-slate-800">Let's set up your project</h1>

<div class="flex flex-col gap-y-4">
  <h2 class="text-subheading">
    Choose an OLAP database for modeling data and serving dashboards.
    <a
      href="https://docs.rilldata.com/concepts/OLAP"
      target="_blank"
      rel="noopener noreferrer">Learn more</a
    >
  </h2>

  <div>
    <PickOlapManagement {managementType} {onSelectManagementType} />
    <PickOlap
      {managementType}
      selectedOLAP={olapDriver}
      {onSelectOLAP}
      on:continue={onContinue}
    />
  </div>
</div>

{#if managementType === "rill-managed"}
  <PickFirstSource
    {olapDriver}
    {firstDataSource}
    {onSelectFirstDataSource}
    {onContinue}
    onSkip={handleSkipFirstSource}
  />
{/if}
