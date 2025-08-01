<script lang="ts">
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";

  export let gitRemote: string;

  const repoName = getRepoNameFromGitRemote(gitRemote);
</script>

<a
  href={getGitUrlFromRemote(gitRemote)}
  target="_blank"
  class="font-semibold font-mono text-gray-800 text-[16px] leading-5"
  rel="noreferrer noopener"
>
  {repoName}
</a>
