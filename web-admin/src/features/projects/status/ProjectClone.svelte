<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: gitRemote = $proj.data?.project?.gitRemote;
  $: managedGitId = $proj.data?.project?.managedGitId;
  $: isGithubConnected = !!gitRemote && !managedGitId;
  $: githubUrl = gitRemote ? getGitUrlFromRemote(gitRemote) : "";

  // CLI commands
  $: cloneCommand = `rill project clone ${project}`;
  $: rillStartCommand = `rill start ${githubUrl}.git`;
  $: envPullCommand = `rill env pull --project ${project}`;

  let copiedCommand: string | null = null;

  function onCopy(command: string) {
    copyToClipboard(command, "Command copied to clipboard", false);
    copiedCommand = command;
    setTimeout(() => {
      copiedCommand = null;
    }, 2000);
  }
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[400px]">
    <span
      class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >
      Local Development
    </span>

    <span class="my-1">
      Clone this project to develop locally.
      <a
        href="https://docs.rilldata.com/developers/guides/clone-a-project"
        target="_blank"
        class="text-primary-600"
      >
        Learn more ->
      </a>
    </span>

    <div class="flex flex-col gap-y-2">
      {#if isGithubConnected}
        <button
          class="command-box"
          title={rillStartCommand}
          on:click={() => onCopy(rillStartCommand)}
        >
          <code class="text-xs truncate">{rillStartCommand}</code>
          <span class="text-gray-400">
            {#if copiedCommand === rillStartCommand}
              <Check size="14px" color="#22c55e" />
            {:else}
              <CopyIcon size="14px" />
            {/if}
          </span>
        </button>

        <div class="env-note">
          <span class="text-[11px] text-gray-500">
            Then pull environment variables:
          </span>
          <button
            class="command-box"
            title={envPullCommand}
            on:click={() => onCopy(envPullCommand)}
          >
            <code class="text-[11px] truncate">{envPullCommand}</code>
            <span class="text-gray-400">
              {#if copiedCommand === envPullCommand}
                <Check size="14px" color="#22c55e" />
              {:else}
                <CopyIcon size="14px" />
              {/if}
            </span>
          </button>
        </div>
      {:else}
        <button
          class="command-box"
          title={cloneCommand}
          on:click={() => onCopy(cloneCommand)}
        >
          <code class="text-xs truncate">{cloneCommand}</code>
          <span class="text-gray-400">
            {#if copiedCommand === cloneCommand}
              <Check size="14px" color="#22c55e" />
            {:else}
              <CopyIcon size="14px" />
            {/if}
          </span>
        </button>
      {/if}
    </div>
  </div>
{/if}

<style lang="postcss">
  .command-box {
    @apply flex items-center justify-between gap-x-2;
    @apply bg-gray-50 border border-gray-200 rounded px-2 py-1;
    @apply font-mono text-gray-800 text-left;
    @apply cursor-pointer w-full;
  }

  .env-note {
    @apply flex flex-col gap-y-1 mt-1 pt-2 border-t border-gray-100;
  }
</style>
