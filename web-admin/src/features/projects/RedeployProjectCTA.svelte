<script lang="ts">
  import {
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
    type V1GetProjectResponse,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import MoonCircleOutline from "@rilldata/web-common/components/icons/MoonCircleOutline.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  const redeployMutation = createAdminServiceRedeployProject();

  // Track waking state: true when mutation is pending OR has succeeded (waiting for refetch)
  $: isWaking = $redeployMutation.isPending || $redeployMutation.isSuccess;

  const REFETCH_INTERVAL = 2000;

  async function handleWakeProject() {
    try {
      await $redeployMutation.mutateAsync({
        org: organization,
        project: project,
      });

      while (true) {
        await queryClient.refetchQueries({
          queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          exact: true,
        });
        const projectQueryResp = queryClient.getQueryData<V1GetProjectResponse>(
          getAdminServiceGetProjectQueryKey(organization, project),
        );
        // If there is a deployment, project refetch logic will handle the rest.
        if (projectQueryResp.deployment) {
          break;
        }
        // Refetch until a deployment is created.
        await new Promise((resolve) => setTimeout(resolve, REFETCH_INTERVAL));
      }
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: m.project_failed_to_wake({ error: getRpcErrorMessage(err) }),
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
            ? m.project_waking_up()
            : m.project_is_hibernating()}
        </CtaHeader>
        <Button
          type="primary"
          wide
          disabled={isWaking}
          loading={isWaking}
          loadingCopy={m.project_waking()}
          onClick={handleWakeProject}
        >
          {m.project_wake()}
        </Button>
        <CtaNeedHelp />
      </svelte:fragment>
      <svelte:fragment slot="read-project">
        <MoonCircleOutline
          size="104px"
          className="text-gray-300"
          gradientStopColor="slate-200"
        />
        <CtaHeader variant="bold">{m.project_this_is_hibernating()}</CtaHeader>
        <CtaMessage>
          {m.project_contact_admin_to_redeploy()}
        </CtaMessage>
        <CtaNeedHelp />
      </svelte:fragment>
    </ProjectAccessControls>
  </CtaContentContainer>
</CtaLayoutContainer>
