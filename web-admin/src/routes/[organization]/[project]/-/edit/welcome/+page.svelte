<script lang="ts">
  import { page } from "$app/state";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { projectWelcomeStatusStores } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import { checkpointProject } from "@rilldata/web-admin/features/projects/publish-project.ts";

  const runtimeClient = useRuntimeClient();

  let project = $derived(page.params.project);

  async function handleDone() {
    projectWelcomeStatusStores.setProjectWelcomeStep(project, false);
    await checkpointProject(runtimeClient);
  }
</script>

<div class="my-auto">
  <TitleContent />

  <div class="flex flex-col py-6 gap-[28px]">
    <div class="flex flex-col mx-auto md:flex-row gap-x-12 gap-y-6">
      <ConnectYourDataWidget onWelcomeScreen />
    </div>

    <p class="text-base font-normal text-fg-secondary text-center">
      Or jump right into an example project.
    </p>

    <ProjectCards skipNavigation onSelect={handleDone} />
  </div>
</div>
