<script lang="ts">
  import { goto } from "$app/navigation";
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";

  export let data;

  const runtimeClient = useRuntimeClient();

  let addDataStep: AddDataStep = AddDataStep.SelectConnector;

  $: isImportStep = addDataStep === AddDataStep.Import;

  async function handleDone() {
    // Push the initial commit to the current branch.
    await runtimeServiceGitPush(runtimeClient, {
      commitMessage: "Initial dashboard commit",
    });

    return goto(`/${data.organization.name}/${data.project.name}`);
  }
</script>

<div class="my-auto">
  {#if !isImportStep}
    <div class="text-base font-semibold text-fg-secondary">Getting started</div>
    <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  {/if}
  <div class="w-fit h-fit mt-4">
    <AddDataManager
      config={{ welcomeScreen: true, skipNavigation: true }}
      initSchema={data.schema}
      onStepChange={(step) => (addDataStep = step)}
      onClose={() => window.history.back()}
      onDone={handleDone}
    />
  </div>
</div>
