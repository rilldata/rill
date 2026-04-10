<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import { runtimeServiceGitPush } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { DeployingDashboardUrlParam } from "@rilldata/web-common/features/project/deploy/utils.ts";

  const runtimeClient = useRuntimeClient();

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  async function handleDone(dashboardName?: string) {
    // Push the initial commit to the current branch.
    await runtimeServiceGitPush(runtimeClient, {
      commitMessage: "Initial dashboard commit",
    });

    setTimeout(
      () =>
        void goto(
          `/${organization}/${project}/-/deploying?${DeployingDashboardUrlParam}=${dashboardName}`,
        ),
      50,
    );
  }
</script>

<div class="my-auto">
  <TitleContent />

  <div class="flex flex-col py-6 gap-[28px]">
    <div class="flex flex-col mx-auto md:flex-row gap-x-12 gap-y-6">
      <ConnectYourDataWidget
        pathPrefix={`/${organization}/${project}/-`}
        onWelcomeScreen
      />
    </div>

    <p class="text-base font-normal text-fg-secondary text-center">
      Or jump right into an example project.
    </p>

    <ProjectCards skipNavigation onSelect={handleDone} allowEmpty={false} />
  </div>
</div>
