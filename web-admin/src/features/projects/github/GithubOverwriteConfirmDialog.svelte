<script lang="ts">
  import { getRepoNameFromGitRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let open = false;
  export let loading: boolean;
  export let error: string | undefined;

  export let onConfirm: () => Promise<void>;
  export let onCancel: () => void = () => {};
  export let githubRemote: string;
  export let subpath: string;

  $: path =
    `${getRepoNameFromGitRemote(githubRemote)}` +
    (subpath ? `/${subpath}` : "");

  const CONFIRMATION_TEXT = "overwrite";

  let confirmInput = "";
  $: confirmed = confirmInput === CONFIRMATION_TEXT;

  function close() {
    onCancel();
    open = false;
  }

  async function handleContinue() {
    await onConfirm();
  }
</script>

<AlertDialog
  bind:open
  onOpenChange={(o) => {
    if (o) {
      confirmInput = "";
    }
  }}
>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle class="flex flex-row gap-x-2 items-center">
        <AlertCircleOutline size="40px" className="text-yellow-600" />
        <div>{m.github_pull_changes_from({ path })}</div>
      </AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-1.5">
        <div>
          {m.github_overwrite_warning()}
        </div>
        <div class="mt-1">
          {m.github_type_to_confirm({ text: CONFIRMATION_TEXT })}
        </div>
        <Input bind:value={confirmInput} id="confirmation" label="" />
        {#if error}
          <div class="text-red-500 text-sm py-px">{error}</div>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button type="secondary" onClick={close}>{m.common_cancel()}</Button>
      <Button
        type="primary"
        onClick={handleContinue}
        disabled={!confirmed}
        {loading}
      >
        {m.common_continue()}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
