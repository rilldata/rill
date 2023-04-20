<script lang="ts">
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected = !!$proj.data?.project?.githubUrl;
  $: prettyGithubRepo = $proj.data?.project?.githubUrl?.split("github.com/")[1];
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1">
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
            class="flex items-center gap-x-1 text-gray-800"
            target="_blank"
            rel="noreferrer"
          >
            <Github className="inline-block w-4 h-4" />
            <span class="font-semibold text-[12px] leading-5 font-mono"
              >{prettyGithubRepo}</span
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
