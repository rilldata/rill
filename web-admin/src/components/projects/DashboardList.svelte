<script lang="ts">
  import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
  import DashboardIcon from "@rilldata/web-common/components/icons/DashboardIcon.svelte";
  import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
    V1GetProjectResponse,
  } from "../../client";

  export let organization: string;
  export let project: string;

  let dashboards: V1MetricsView[];

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj?.isSuccess && $proj.data?.prodDeployment) {
    updateDashboardsForProject($proj.data);
  }

  // This method has to be here since we cannot have async-await in reactive statement to set dashboardListItems
  async function updateDashboardsForProject(projectData: V1GetProjectResponse) {
    const status = projectData.prodDeployment.status;
    if (status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING) return;

    dashboards = await getDashboardsForProject(projectData);
  }
</script>

{#if dashboards?.length === 0}
  <p class="text-gray-500 text-xs">This project has no dashboards yet.</p>
{:else if dashboards?.length > 0}
  <ol class="flex flex-col gap-y-4 max-w-full 2xl:max-w-[1200px]">
    {#each dashboards as dashboard}
      <li class="w-full h-[52px] border rounded">
        <a
          href={`/${organization}/${project}/${dashboard.name}`}
          class="w-full h-full overflow-x-auto p-3 flex items-center gap-x-6 text-gray-700 hover:text-blue-600 hover:bg-slate-50"
        >
          <!-- Icon -->
          <div
            class="ml-1 w-4 h-4 inline-flex justify-center items-center text-gray-400"
          >
            <DashboardIcon size={"14px"} />
          </div>

          <div class="flex items-center gap-x-8">
            <!-- Name -->
            <span
              class="text-sm font-medium shrink-0 truncate"
              title={dashboard?.label || dashboard.name}
            >
              {dashboard?.label || dashboard.name}
            </span>

            <!-- We'll show errored dashboards again when we integrate the new Reconcile -->
            <!-- Error tag -->
            <!-- {#if $proj.data.prodDeployment.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING && !dashboard.isValid}
              <Tooltip distance={8} location="right">
                <TooltipContent slot="tooltip-content">
                  <ProjectAccessControls {organization} {project}>
                    <svelte:fragment slot="manage-project">
                      <span class="text-xs">
                        This dashboard has an error. Please check the project
                        logs.
                      </span>
                    </svelte:fragment>
                    <svelte:fragment slot="read-project">
                      <span class="text-xs">
                        This dashboard has an error. Please contact the project
                        administrator.
                      </span>
                    </svelte:fragment>
                  </ProjectAccessControls>
                </TooltipContent>
                <Tag>Error</Tag>
              </Tooltip>
            {/if} -->

            <!-- Description -->
            {#if dashboard.description}
              <!-- Note: line-clamp-2 uses `display: -webkit-box;` which overrides the `hidden` class -->
              <span
                class="text-gray-800 text-xs font-light break-normal hidden sm:line-clamp-2"
                >{dashboard.description}</span
              >
            {/if}
          </div>
        </a>
      </li>
    {/each}
  </ol>
{/if}
