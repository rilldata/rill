<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";

  let open = false;

  export let organization: string;
  export let project: string;
  export let gitRemote: string | undefined = undefined;
  export let managedGitId: string | undefined = undefined;

  $: githubUrl = gitRemote ? getGitUrlFromRemote(gitRemote) : "";
  $: isGithubConnected = !!gitRemote && !managedGitId && !!githubUrl;

  // CLI commands
  $: cloneCommand = `rill project clone --org ${organization} ${project}`;
  $: rillStartCommand = `rill start ${githubUrl}.git`;
  let copiedCommand: string | null = null;

  function onCopy(command: string) {
    copyToClipboard(command, "Command copied to clipboard", false);
    copiedCommand = command;
    setTimeout(() => {
      copiedCommand = null;
    }, 2000);
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger asChild let:builder>
    <Button type="secondary" builders={[builder]}>Download Project</Button>
  </Popover.Trigger>

  <Popover.Content class="w-[380px]" align="end" sideOffset={8}>
    <div class="flex flex-col gap-y-3">
      <span class="text-sm text-fg-secondary">
        Clone this project to develop locally.
        <a
          href="https://docs.rilldata.com/developers/tutorials/clone-a-project"
          target="_blank"
          rel="noopener noreferrer"
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
            <span class="text-fg-muted">
              {#if copiedCommand === rillStartCommand}
                <Check size="14px" color="#22c55e" />
              {:else}
                <CopyIcon size="14px" />
              {/if}
            </span>
          </button>
        {:else}
          <button
            class="command-box"
            title={cloneCommand}
            on:click={() => onCopy(cloneCommand)}
          >
            <code class="text-xs truncate">{cloneCommand}</code>
            <span class="text-fg-muted">
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
  </Popover.Content>
</Popover.Root>

<style lang="postcss">
  .command-box {
    @apply flex items-center justify-between gap-x-2;
    @apply bg-surface-subtle border border-gray-200 rounded px-2 py-1;
    @apply font-mono text-fg-primary text-left;
    @apply cursor-pointer w-full;
  }

  .command-box:hover {
    @apply bg-surface-hover;
  }
</style>
