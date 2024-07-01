<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import { isProjectPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectDashboardsListener from "@rilldata/web-admin/features/projects/ProjectDashboardsListener.svelte";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { useProjectDeployment } from "@rilldata/web-admin/features/projects/status/selectors";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import ProjectTabs from "../../../features/projects/ProjectTabs.svelte";
  import { useProjectRuntime } from "../../../features/projects/selectors";
  import { viewAsUserStore } from "../../../features/view-as-user/viewAsUserStore";

  const user = createAdminServiceGetCurrentUser();

  $: ({ organization, project } = $page.params);

  $: projRuntime = useProjectRuntime(organization, project);
  $: ({ data: runtime, isSuccess: runtimeQueryIsSuccess } = $projRuntime);

  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment } = $projectDeployment);

  $: if (project && $user.data?.user?.id) {
    metricsService.loadCloudFields({
      isDev: window.location.host.startsWith("localhost"),
      projectId: project,
      organizationId: organization,
      userId: $user.data?.user?.id,
    });
  }

  $: onProjectPage = isProjectPage($page);
</script>

{#if onProjectPage && deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_OK}
  <ProjectTabs />
{/if}

{#if $viewAsUserStore}
  <!-- When the user is being spoofed via the "View As" functionality, we don't provide the runtime here.
In these cases, the "View as" actions manually set the runtime.  -->
  <slot />
{:else if !deployment}
  <!-- No deployment = the project is "hibernating" -->
  <RedeployProjectCta {organization} {project} />
{:else if deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
  <ProjectBuilding />
{:else if deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR}
  <ErrorPage
    statusCode={500}
    header="Deployment Error"
    body={deployment?.statusMessage !== ""
      ? deployment?.statusMessage
      : "There was an error deploying your project. Please contact support."}
  />
{:else if deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_OK && runtimeQueryIsSuccess}
  <RuntimeProvider
    host={runtime?.host}
    instanceId={runtime?.instanceId}
    jwt={runtime?.jwt}
  >
    <ProjectDashboardsListener>
      <slot />
    </ProjectDashboardsListener>
  </RuntimeProvider>
{/if}
