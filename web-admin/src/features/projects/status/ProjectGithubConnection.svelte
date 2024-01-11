<script lang="ts">
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceGetProject } from "../../../client";
  import { useDashboardsLastUpdated } from "../../dashboards/listing/selectors";
  import { getRepoNameFromGithubUrl } from "../github-utils";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected = !!$proj.data?.project?.githubUrl;
  $: repoName =
    $proj.data?.project?.githubUrl &&
    getRepoNameFromGithubUrl($proj.data.project.githubUrl);
  $: subpath = $proj.data?.project?.subpath;
  $: githubLastSynced = useDashboardsLastUpdated(
    $runtime.instanceId,
    organization,
    project,
  );
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
      >Github</span
    >
    {#if isGithubConnected}
      <div class="flex items-start gap-x-1">
        <div class="py-0.5">
          <Github className="w-4 h-4" />
        </div>
        <div class="flex flex-col">
          <a
            href={$proj.data?.project?.githubUrl}
            class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
            target="_blank"
            rel="noreferrer"
          >
            {repoName}
          </a>
          {#if subpath}
            <div class="flex items-center">
              <span class="font-mono">subpath</span>
              <span class="text-gray-800">
                : /{subpath}
              </span>
            </div>
          {/if}
          {#if $githubLastSynced}
            <span class="text-gray-500 text-[11px] leading-4">
              Synced {$githubLastSynced.toLocaleString(undefined, {
                month: "short",
                day: "numeric",
                hour: "numeric",
                minute: "numeric",
              })}
            </span>
          {/if}
        </div>
      </div>
    {:else}
      <Tooltip alignment="start" distance={4}>
        <div class="flex items-center gap-x-1">
          <AlertCircleOutline className="text-red-400" size={"16px"} />
          <span>Not connected to a repository</span>
        </div>
        <TooltipContent slot="tooltip-content" maxWidth="300px"
          >This project is no longer connected to a Github repository. This is
          how we ensure this project is up to date.</TooltipContent
        >
      </Tooltip>
    {/if}
  </div>
{/if}
