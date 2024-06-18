<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { isProjectPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectDashboardsListener from "@rilldata/web-admin/features/projects/ProjectDashboardsListener.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import ProjectTabs from "../../../features/projects/ProjectTabs.svelte";
  import { useProjectRuntime } from "../../../features/projects/selectors";
  import { viewAsUserStore } from "../../../features/view-as-user/viewAsUserStore";

  $: ({ organization, project } = $page.params);

  $: projRuntime = useProjectRuntime(organization, project);
  $: ({ data: runtime } = $projRuntime);

  const user = createAdminServiceGetCurrentUser();

  $: isRuntimeHibernating = $projRuntime.isSuccess && !$projRuntime.data;

  $: if (isRuntimeHibernating) {
    // Redirect any nested routes (notably dashboards) to the project page
    goto(`/${organization}/${project}`);
  }

  $: onProjectPage = isProjectPage($page);

  $: if (project && $user.data?.user?.id) {
    metricsService.loadCloudFields({
      isDev: window.location.host.startsWith("localhost"),
      projectId: project,
      organizationId: organization,
      userId: $user.data?.user?.id,
    });
  }
</script>

{#if $viewAsUserStore}
  <!-- When the user is being spoofed via the "View As" functionality, we don't provide the runtime here.
    In these cases, the "View as" actions manually set the runtime.  -->
  <slot />
{:else if isRuntimeHibernating}
  <!-- When the runtime is hibernating, we omit the RuntimeProvider. -->
  <slot />
{:else}
  <RuntimeProvider
    host={runtime?.host}
    instanceId={runtime?.instanceId}
    jwt={runtime?.jwt}
  >
    <ProjectDashboardsListener>
      <!-- We make sure to put the project tabs within the `RuntimeProvider` so we can add decoration 
        to the tab labels that query the runtime (e.g. the project status badge) -->
      {#if onProjectPage}
        <ProjectTabs />
      {/if}
      <slot />
    </ProjectDashboardsListener>
  </RuntimeProvider>
{/if}
