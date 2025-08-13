<script lang="ts">
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import {
    getGitUrlFromRemote,
    getRepoNameFromGitRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils.ts";

  export let gitRemote: string;
  export let subpath: string;
  export let branch: string;

  $: repoName = getRepoNameFromGitRemote(gitRemote);
</script>

<div class="w-fit mx-auto">
  <div class="flex flex-row gap-x-1 items-center">
    <Github className="w-4 h-4" />
    <a
      href={getGitUrlFromRemote(gitRemote)}
      class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
      target="_blank"
      rel="noreferrer noopener"
    >
      {repoName}
    </a>
  </div>
  {#if subpath}
    <div class="flex items-center">
      <span class="font-mono">subpath</span>
      <span class="text-gray-800">
        : /{subpath}
      </span>
    </div>
  {/if}
  <div class="flex items-center">
    <span class="font-mono">branch</span>
    <span class="text-gray-800">
      : {branch}
    </span>
  </div>
</div>
