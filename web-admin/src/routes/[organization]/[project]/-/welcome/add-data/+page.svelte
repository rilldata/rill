<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";
  import type { PageData } from "./$types";
  import { DeployingDashboardUrlParam } from "@rilldata/web-common/features/project/deploy/utils.ts";
  import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";

  let { data }: { data: PageData } = $props();

  const runtimeClient = useRuntimeClient();

  let addDataStep = $state<AddDataStep>(AddDataStep.SelectConnector);

  let isImportStep = $derived(addDataStep === AddDataStep.Import);

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  async function handleDone(generatedDashboard?: string) {
    // Push the initial commit to the current branch.
    await runtimeServiceGitPush(runtimeClient, {
      commitMessage: "Initial dashboard commit",
    });

    setTimeout(
      () =>
        void goto(
          `/${organization}/${project}/-/deploying?${DeployingDashboardUrlParam}=${generatedDashboard}`,
        ),
      50,
    );
  }

  onMount(async () => {
    // Prefetch connectors and load into cache.
    await fetchAnalyzeConnectors(runtimeClient);
  });
</script>

<div class="my-auto">
  {#if !isImportStep}
    <div class="text-base font-semibold text-fg-secondary">Getting started</div>
    <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  {/if}
  <div class="w-fit h-fit mt-4">
    {#key data.schema}
      <AddDataManager
        config={{
          welcomeScreen: true,
          skipNavigation: true,
          pathPrefix: `/${organization}/${project}/-`,
        }}
        initSchema={data.schema}
        onStepChange={(step) => (addDataStep = step)}
        onClose={() => window.history.back()}
        onDone={handleDone}
      />
    {/key}
  </div>
</div>
