<script lang="ts">
  import { createAdminServiceConnectProjectToGithub } from "@rilldata/web-admin/client";
  import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors.ts";
  import { GithubAccessManager } from "@rilldata/web-admin/features/projects/github/GithubAccessManager.ts";
  import { getGithubUserOrgs } from "@rilldata/web-admin/features/projects/github/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let organization: string;
  export let project: string;

  const FORM_ID = "github-connect-form";

  const githubAccessManager = new GithubAccessManager();
  const { githubConnectionFailed, userStatus } = githubAccessManager;

  const connectProjectToGithub = createAdminServiceConnectProjectToGithub();
  $: ({ error, isPending } = $connectProjectToGithub);
  $: parsedError = error ? extractGithubConnectError(error as any) : null;

  const githubUserOrgs = getGithubUserOrgs();

  const initialValues: {
    org: string;
    name: string;
  } = {
    org: "",
    name: project, // Initialize repo name with project name
  };
  const schema = yup(
    object({
      org: string().required("Org is required"),
      name: string().required("Repo name is required"),
    }),
  );

  const { form, errors, enhance, submit, reset } = superForm(
    defaults(initialValues, schema),
    {
      id: FORM_ID,
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values = form.data;

        const remote = `https://github.com/${values.org}/${values.name}.git`;

        await $connectProjectToGithub.mutateAsync({
          project: project,
          organization: organization,
          data: {
            remote,
            create: true,
          },
        });
      },
    },
  );

  $: disableSubmit = isPending || $githubConnectionFailed;
</script>

<Dialog.Root
  bind:open
  onOpenChange={(o) => {
    if (!o) reset();
  }}
>
  <Dialog.Trigger asChild let:builder>
    <Button
      builders={[builder]}
      type="primary"
      class="w-fit mt-1"
      loading={$userStatus.isFetching}
      onClick={() => void githubAccessManager.ensureGithubAccess()}
    >
      <Github className="w-4 h-4" fill="white" />
      Connect to GitHub
    </Button>
  </Dialog.Trigger>
  <Dialog.Content class="translate-y-[-200px]">
    <Dialog.Header>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <div class="flex flex-col gap-y-1">
          <Dialog.Title>Connect to github</Dialog.Title>
          <Dialog.Description>
            Connect this project to a new repo.
          </Dialog.Description>
        </div>
      </div>
    </Dialog.Header>

    <form
      id={FORM_ID}
      on:submit|preventDefault={submit}
      use:enhance
      class="flex flex-col gap-y-3"
    >
      <Select
        bind:value={$form.org}
        id="org"
        label="Repository org"
        options={$githubUserOrgs.data ?? []}
        optionsLoading={$githubUserOrgs.isFetching}
        sameWidth
        enableSearch
        onAddNew={() => githubAccessManager.reselectOrgs()}
        addNewLabel="+ Connect other orgs"
      />

      <Input
        bind:value={$form.name}
        errors={$errors?.name}
        id="name"
        label="Repository name"
        capitalizeLabel={false}
      />

      {#if parsedError?.message}
        <div class="text-red-500 text-sm py-px">
          {parsedError.message}
        </div>
      {/if}

      {#if $githubConnectionFailed}
        <div class="text-red-500 text-sm py-px">
          <div>Failed to connect to GitHub. Please try again.</div>
          <Button
            type="secondary"
            onClick={() => githubAccessManager.ensureGithubAccess()}
          >
            Reconnect
          </Button>
        </div>
      {/if}
    </form>

    <Dialog.Footer>
      <Button
        onClick={() => {
          open = false;
          reset();
        }}
        type="secondary"
      >
        Cancel
      </Button>
      <Button
        form={FORM_ID}
        submitForm
        type="primary"
        loading={disableSubmit}
        disabled={disableSubmit}
      >
        Create and push changes
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
