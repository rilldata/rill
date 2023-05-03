<script lang="ts">
  export let githubUsername: string;

  const githubUrl = `https://github.com/${githubUsername}`;
</script>

<a
  href={githubUrl}
  target="_blank"
  class="font-semibold font-mono text-gray-800 text-[16px] leading-5"
  rel="noreferrer"
>
  {githubUsername}
</a>
