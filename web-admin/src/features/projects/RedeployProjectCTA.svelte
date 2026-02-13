<script lang="ts">
  import {
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import MoonCircleOutline from "@rilldata/web-common/components/icons/MoonCircleOutline.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  const redeployMutation = createAdminServiceRedeployProject();

  async function handleWakeProject() {
    try {
      await $redeployMutation.mutateAsync({
        org: organization,
        project: project,
      });

      void queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
        exact: true,
      });

      eventBus.emit("notification", {
        type: "success",
        message: "Project is now waking up",
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to wake project: ${getRpcErrorMessage(err)}`,
      });
    }
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <div class="relative size-[104px]">
          <div
            class="absolute inset-0 transition-opacity duration-200"
            class:opacity-0={$redeployMutation.isPending}
          >
            <MoonCircleOutline
              size="104px"
              className="text-gray-300"
              gradientStopColor="slate-200"
            />
          </div>
          <div
            class="absolute inset-0 transition-opacity duration-200"
            class:opacity-0={!$redeployMutation.isPending}
          >
            <LoadingCircleOutline size="104px" className="text-gray-300" />
          </div>
        </div>
        <CtaHeader variant="bold">
          {$redeployMutation.isPending
            ? "Waking up your project..."
            : "Your project is hibernating"}
        </CtaHeader>
        <Button
          type="primary"
          wide
          loading={$redeployMutation.isPending}
          loadingCopy="Waking..."
          onClick={handleWakeProject}
        >
          Wake project
        </Button>
        <div
          class="flex flex-col items-center gap-y-6 transition-opacity duration-200"
          class:opacity-40={$redeployMutation.isPending}
          class:pointer-events-none={$redeployMutation.isPending}
        >
          <CtaMessage>
            You can also run the following command in the Rill CLI:
          </CtaMessage>
          <CLICommandDisplay
            command="rill project hibernate {project} --redeploy"
          />
        </div>
        <CtaNeedHelp />
      </svelte:fragment>
      <svelte:fragment slot="read-project">
        <MoonCircleOutline
          size="104px"
          className="text-gray-300"
          gradientStopColor="slate-200"
        />
        <CtaHeader variant="bold">This project is hibernating</CtaHeader>
        <CtaMessage>
          Contact the project's administrator to redeploy the project.
        </CtaMessage>
        <CtaNeedHelp />
      </svelte:fragment>
    </ProjectAccessControls>
  </CtaContentContainer>
</CtaLayoutContainer>
