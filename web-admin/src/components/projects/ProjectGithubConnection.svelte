<script lang="ts">
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetProject } from "../../client";
  import { getRepoNameFromGithubUrl } from "./github-utils";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected = !!$proj.data?.project?.githubUrl;
  $: repoName =
    $proj.data?.project?.githubUrl &&
    getRepoNameFromGithubUrl($proj.data.project.githubUrl);
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[340px]">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
      >Github</span
    >
    <div>
      {#if isGithubConnected}
        <div class="flex items-center gap-x-1">
          <CheckCircle className="text-blue-500" size={"16px"} />
          <span>Connected to </span>
          <a
            href={$proj.data?.project?.githubUrl}
            class="flex items-center gap-x-1 text-gray-800 flex-1 truncate"
            target="_blank"
            rel="noreferrer"
          >
            <Github className="inline-block w-4 h-4" />
            <span class="font-semibold text-[12px] leading-5 font-mono truncate"
              >{repoName}</span
            ></a
          >
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
  </div>
{/if}
