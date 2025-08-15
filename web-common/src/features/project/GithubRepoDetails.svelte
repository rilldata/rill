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

<div class="flex flex-col w-fit mx-auto text-center text-base text-sm">
  <div class="flex flex-row gap-x-1 items-center">
    <Github className="w-4 h-4" />
    <a
      href={getGitUrlFromRemote(gitRemote)}
      class="text-gray-800 font-semibold font-mono truncate"
      target="_blank"
      rel="noreferrer noopener"
    >
      {repoName}
    </a>
  </div>
  {#if subpath}
    <div>
      <span>subpath:</span>
      <span class="text-gray-800 font-mono">
        /{subpath}
      </span>
    </div>
  {/if}
  <div>
    <span>branch:</span>
    <span class="text-gray-800 font-mono">
      {branch}
    </span>
  </div>
</div>
