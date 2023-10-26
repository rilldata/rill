<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ProjectDashboardsListener from "@rilldata/web-admin/features/projects/ProjectDashboardsListener.svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { isProjectPage } from "../../../features/navigation/nav-utils";
  import ProjectTabs from "../../../features/projects/ProjectTabs.svelte";
  import { useProjectRuntime } from "../../../features/projects/selectors";
  import { viewAsUserStore } from "../../../features/view-as-user/viewAsUserStore";

  $: projRuntime = useProjectRuntime(
    $page.params.organization,
    $page.params.project
  );

  $: isRuntimeHibernating = $projRuntime.isSuccess && !$projRuntime.data;

  $: if (isRuntimeHibernating) {
    // Redirect any nested routes (notably dashboards) to the project page
    goto(`/${$page.params.organization}/${$page.params.project}`);
  }

  $: onProjectPage = isProjectPage($page);
</script>

<!-- Note: we don't provide the runtime here when the user is being spoofed via the "View As" functionality.
    In these cases, the "View as" actions manually set the runtime.  -->
{#if !$viewAsUserStore}
  <RuntimeProvider
    host={$projRuntime.data?.host}
    instanceId={$projRuntime.data?.instanceId}
    jwt={$projRuntime.data?.jwt}
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
{:else}
  <slot />
{/if}
