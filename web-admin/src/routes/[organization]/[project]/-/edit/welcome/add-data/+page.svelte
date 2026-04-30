<script lang="ts">
  import { page } from "$app/state";
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { projectWelcomeStatus } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import { checkpointProject } from "@rilldata/web-admin/features/projects/publish-project.ts";
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  const runtimeClient = useRuntimeClient();

  let addDataStep = $state<AddDataStep>(AddDataStep.SelectConnector);

  let isImportStep = $derived(addDataStep === AddDataStep.Import);

  let project = $derived(page.params.project);

  // Prefetch connectors and load into cache. We will show a spinner while this is fetching.
  let connectorsQuery = $derived(
    createRuntimeServiceAnalyzeConnectors(runtimeClient, {}),
  );

  async function handleDone() {
    projectWelcomeStatus.setProjectWelcomeStep(project, false);
    await checkpointProject(runtimeClient);
  }
</script>

<div class="my-auto">
  {#if !isImportStep}
    <div class="text-base font-semibold text-fg-secondary">Getting started</div>
    <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  {/if}
  <div class="w-fit h-fit mt-4">
    {#if $connectorsQuery.isPending}
      <Spinner status={EntityStatus.Running} size="3rem" duration={725} />
    {:else if $connectorsQuery.data}
      {#key data.schema}
        <AddDataManager
          config={{
            welcomeScreen: true,
          }}
          initSchema={data.schema}
          onStepChange={(step) => (addDataStep = step)}
          onClose={() => window.history.back()}
          onDone={handleDone}
        />
      {/key}
    {/if}
  </div>
</div>
