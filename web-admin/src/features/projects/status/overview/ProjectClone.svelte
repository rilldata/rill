<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CopyableCodeBlock from "@rilldata/web-common/components/calls-to-action/CopyableCodeBlock.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
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
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button {...props} type="secondary">Download Project</Button>
    {/snippet}
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
          <CopyableCodeBlock code={rillStartCommand} />
        {:else}
          <CopyableCodeBlock code={cloneCommand} />
        {/if}
      </div>
    </div>
  </Popover.Content>
</Popover.Root>
