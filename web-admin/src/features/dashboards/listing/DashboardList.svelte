<script lang="ts">
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import DashboardIcon from "@rilldata/web-common/components/icons/DashboardIcon.svelte";
  import { Tag } from "@rilldata/web-common/components/tag";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetProject } from "../../../client";

  export let organization: string;
  export let project: string;

  let dashboards: ReturnType<typeof useDashboards>;

  $: proj = createAdminServiceGetProject(organization, project);
  $: if ($proj?.isSuccess && $proj.data?.prodDeployment) {
    dashboards = useDashboards($proj.data.prodDeployment.runtimeInstanceId);
  }
</script>

{#if $dashboards?.data?.length === 0}
  <p class="text-gray-500 text-xs">This project has no dashboards yet.</p>
{:else if $dashboards?.data?.length > 0}
  <ol class="flex flex-col gap-y-4 max-w-full 2xl:max-w-[1200px]">
    {#each $dashboards.data as dashboard}
      <li class="w-full h-[52px] border rounded">
        <a
          href={`/${organization}/${project}/${dashboard.meta.name.name}`}
          class="w-full h-full overflow-x-auto p-3 flex items-center gap-x-6 text-gray-700 hover:text-primary-600 hover:bg-slate-50"
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
              title={dashboard.metricsView?.state?.validSpec?.title ||
                dashboard.meta.name.name}
            >
              {dashboard.metricsView?.state?.validSpec?.title ||
                dashboard.meta.name.name}
            </span>

            <!-- We'll show errored dashboards again when we integrate the new Reconcile -->
            <!-- Error tag -->
            {#if !dashboard.metricsView?.state?.validSpec}
              <Tooltip distance={8} location="right">
                <TooltipContent slot="tooltip-content">
                  <ProjectAccessControls {organization} {project}>
                    <svelte:fragment slot="manage-project">
                      <span class="text-xs">
                        This dashboard has an error. Please check the project
                        status.
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
            {/if}

            <!-- Description -->
            {#if dashboard.metricsView?.state?.validSpec?.description}
              <!-- Note: line-clamp-2 uses `display: -webkit-box;` which overrides the `hidden` class -->
              <span
                class="text-gray-800 text-xs font-light break-normal hidden sm:line-clamp-2"
                >{dashboard.metricsView?.state?.validSpec?.description}</span
              >
            {/if}
          </div>
        </a>
      </li>
    {/each}
  </ol>
{/if}
