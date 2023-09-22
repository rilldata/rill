<script lang="ts">
  import { getRepoNameFromGithubUrl } from "./github-utils";

  export let githubUrl: string;

  const repoName = getRepoNameFromGithubUrl(githubUrl);
</script>

<a
  href={githubUrl}
  target="_blank"
  class="font-semibold font-mono text-gray-800 text-[16px] leading-5"
  rel="noreferrer"
>
  {repoName}
</a>
