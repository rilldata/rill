<script lang="ts">
  import {
    DashboardListItem,
    getDashboardsForProject,
  } from "@rilldata/web-admin/components/projects/dashboards";
  import DashboardIcon from "@rilldata/web-common/components/icons/DashboardIcon.svelte";
  import { Tag } from "@rilldata/web-common/components/tag";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
    V1GetProjectResponse,
  } from "../../client";
  import AccessControls from "../access-controls/AccessControls.svelte";

  export let organization: string;
  export let project: string;

  let dashboardListItems: DashboardListItem[];

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj?.isSuccess && $proj.data?.prodDeployment) {
    updateDashboardsForProject($proj.data);
  }

  // This method has to be here since we cannot have async-await in reactive statement to set dashboardListItems
  async function updateDashboardsForProject(projectData: V1GetProjectResponse) {
    const status = projectData.prodDeployment.status;
    if (status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING) return;

    dashboardListItems = await getDashboardsForProject(projectData);
  }
</script>

{#if dashboardListItems?.length === 0}
  <p class="text-gray-500 text-xs">This project has no dashboards yet.</p>
{:else if dashboardListItems?.length > 0}
  <ol class="flex flex-col gap-y-4 w-fit max-w-full xl:max-w-[1200px]">
    {#each dashboardListItems as dashboardListItem}
      <li
        class="h-[52px] flex items-center gap-x-2 border rounded w-full text-xs font-medium leading-4 break-all"
      >
        <svelte:element
          this={dashboardListItem.isValid ? "a" : "div"}
          href={dashboardListItem.isValid
            ? `/${organization}/${project}/${dashboardListItem.name}`
            : undefined}
          class="w-full h-full p-3 flex items-center gap-x-6 {dashboardListItem.isValid
            ? 'text-gray-700 hover:text-blue-600 hover:bg-slate-50'
            : 'text-gray-400'}"
        >
          <!-- Icon -->
          <div
            class="ml-1 w-4 h-4 inline-flex justify-center items-center text-gray-400"
          >
            <DashboardIcon size={"14px"} />
          </div>

          <div class="flex items-center gap-x-10 w-full">
            <!-- Name -->
            <span class="text-sm font-medium whitespace-nowrap truncate">
              {dashboardListItem?.title || dashboardListItem.name}
            </span>

            <!-- Error tag -->
            {#if !dashboardListItem.isValid}
              <Tooltip distance={8} location="right">
                <TooltipContent slot="tooltip-content">
                  <AccessControls {organization} {project}>
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
                  </AccessControls>
                </TooltipContent>
                <Tag>Error</Tag>
              </Tooltip>
            {/if}

            <!-- Description -->
            {#if dashboardListItem.description}
              <span
                class="text-gray-800 text-xs font-light break-normal line-clamp-2"
                >{dashboardListItem.description}</span
              >
            {/if}
          </div>
        </svelte:element>
      </li>
    {/each}
  </ol>
{/if}
