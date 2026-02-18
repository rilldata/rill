<script lang="ts">
  import {
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import MoonCircleOutline from "@rilldata/web-common/components/icons/MoonCircleOutline.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  const redeployMutation = createAdminServiceRedeployProject();

  // Track waking state: true when mutation is pending OR has succeeded (waiting for refetch)
  $: isWaking = $redeployMutation.isPending || $redeployMutation.isSuccess;

  async function handleWakeProject() {
    try {
      await $redeployMutation.mutateAsync({
  const redeployProjectMutation = createAdminServiceRedeployProject();

  async function wakeProject() {
    try {
      await $redeployProjectMutation.mutateAsync({
        org: organization,
        project: project,
      });

      void queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
        exact: true,
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to wake project: ${getRpcErrorMessage(err)}`,
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });

      eventBus.emit("notification", {
        message: "Project is waking up",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message: axiosError.response?.data?.message ?? "Failed to wake project",
        type: "error",
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
            class:opacity-0={isWaking}
          >
            <MoonCircleOutline
              size="104px"
              className="text-gray-300"
              gradientStopColor="slate-200"
            />
          </div>
          <div
            class="absolute inset-0 transition-opacity duration-200"
            class:opacity-0={!isWaking}
          >
            <LoadingCircleOutline size="104px" className="text-gray-300" />
          </div>
        </div>
        <CtaHeader variant="bold">
          {isWaking
            ? "Waking up your project..."
            : "Your project is hibernating"}
        </CtaHeader>
        <Button
          type="primary"
          wide
          disabled={isWaking}
          loading={isWaking}
          loadingCopy="Waking..."
          onClick={handleWakeProject}
        >
          Wake project
        </Button>
        <CtaNeedHelp />
        <CtaHeader variant="bold">Your project is hibernating</CtaHeader>
        <CtaMessage>
          The project is paused and not consuming resources. Wake the project to
          resume access.
        </CtaMessage>
        <Button
          onClick={wakeProject}
          type="primary"
          loading={$redeployProjectMutation.isPending}
        >
          Wake project
        </Button>
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
