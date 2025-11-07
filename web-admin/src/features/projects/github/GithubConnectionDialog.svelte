<script lang="ts">
  import {
    createAdminServiceConnectProjectToGithub,
    createAdminServiceUpdateProject,
  } from "@rilldata/web-admin/client";
  import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors.ts";
  import { GithubAccessManager } from "@rilldata/web-admin/features/projects/github/GithubAccessManager.ts";
  import GithubOverwriteConfirmDialog from "@rilldata/web-admin/features/projects/github/GithubOverwriteConfirmDialog.svelte";
  import {
    getGithubUserOrgs,
    getGithubUserRepos,
  } from "@rilldata/web-admin/features/projects/github/selectors.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CaretDownFilledIcon from "@rilldata/web-common/components/icons/CaretDownFilledIcon.svelte";
  import CaretRightFilledIcon from "@rilldata/web-common/components/icons/CaretRightFilledIcon.svelte";
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
  const updateProject = createAdminServiceUpdateProject();
  $: error = $connectProjectToGithub.error ?? $updateProject.error;
  $: isPending = $connectProjectToGithub.isPending || $updateProject.isPending;
  $: parsedError = error ? extractGithubConnectError(error as any) : null;

  type GithubConnectType = "create" | "pull";
  const GithubSelectionTypeOptions = [
    {
      label: "Push project to a new GitHub repo",
      value: "create",
    },
    {
      label: "Pull changes from existing GitHub repo",
      value: "pull",
    },
  ];
  let connectType: GithubConnectType = "create";

  let isNewRepoType = false;
  let advancedOpened = false;
  let showOverwriteConfirmation = false;

  const githubUserOrgs = getGithubUserOrgs();
  $: githubUserRepos = getGithubUserRepos(!isNewRepoType);

  const initialValues: {
    type: GithubConnectType;
    repo: string;
    org: string;
    name: string;
  } = {
    type: connectType,
    repo: "",
    org: "",
    name: project, // Initialize repo name with project name
  };
  const schema = yup(
    object({
      type: string().required(),
      org: string().when("type", {
        is: "create",
        then: (schema) => schema.required("Org is required"),
        otherwise: (schema) => schema.notRequired(),
      }),
      name: string().when("type", {
        is: "create",
        then: (schema) => schema.required("Repo name is required"),
        otherwise: (schema) => schema.notRequired(),
      }),
      repo: string().when("type", {
        is: "create",
        then: (schema) => schema.notRequired(),
        otherwise: (schema) => schema.required("Repo is required"),
      }),
      branch: string().when("type", {
        is: "create",
        then: (schema) => schema.notRequired(),
        otherwise: (schema) => schema.required("Branch is required"),
      }),
      subpath: string(),
    }),
  );

  const { form, errors, enhance, submit, reset } = superForm(
    defaults(initialValues, schema),
    {
      id: FORM_ID,
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        console.log(form.valid);
        if (!form.valid) return;
        const values = form.data;

        if (isNewRepoType) {
          const remote = `https://github.com/${values.org}/${values.name}.git`;

          await $connectProjectToGithub.mutateAsync({
            org: organization,
            project,
            data: {
              remote,
            },
          });
        } else {
          await $updateProject.mutateAsync({
            org: organization,
            project,
            data: {
              gitRemote: values.repo,
              prodBranch: values.branch,
              subpath: values.subpath,
            },
          });
        }
      },
    },
  );

  $: isNewRepoType = $form.type === "create";

  $: disableSubmit = isPending || $githubConnectionFailed;

  function onSelectedRepoChange(newRemote: string) {
    const repo = $githubUserRepos.data?.rawRepos?.find(
      (r) => r.remote === newRemote,
    );
    if (!repo?.defaultBranch) return;

    $form.branch = repo.defaultBranch;
  }

  function resetMutations() {
    $connectProjectToGithub.reset();
    $updateProject.reset();
  }
</script>

<Dialog.Root
  bind:open
  onOpenChange={(o) => {
    if (!o) {
      resetMutations();
      reset();
    }
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
      <Github className="w-4 h-4" />
      Connect to GitHub
    </Button>
  </Dialog.Trigger>
  <Dialog.Content class="translate-y-[-200px]">
    <Dialog.Header>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <div class="flex flex-col gap-y-1">
          <Dialog.Title>Connect to GitHub</Dialog.Title>
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
        bind:value={$form.type}
        id="type"
        label="Connection type"
        sameWidth
        options={GithubSelectionTypeOptions}
        onChange={resetMutations}
      />

      {#if isNewRepoType}
        <Select
          bind:value={$form.org}
          id="org"
          label="Repository org"
          options={$githubUserOrgs.data ?? []}
          optionsLoading={$githubUserOrgs.isFetching}
          sameWidth
          enableSearch
          onAddNew={() => githubAccessManager.reselectOrgOrRepos(false)}
          addNewLabel="+ Connect other orgs"
        />

        <Input
          bind:value={$form.name}
          errors={$errors?.name}
          id="name"
          label="Repository name"
          capitalizeLabel={false}
        />
      {:else}
        <Select
          bind:value={$form.repo}
          id="name"
          label="Repository"
          sameWidth
          options={$githubUserRepos.data?.repoOptions ?? []}
          optionsLoading={$githubUserRepos.isFetching}
          enableSearch
          onChange={(newRepo) => onSelectedRepoChange(newRepo)}
          onAddNew={() => githubAccessManager.reselectOrgOrRepos(true)}
          addNewLabel="+ Connect other repos"
        />

        <Collapsible.Root bind:open={advancedOpened}>
          <Collapsible.Trigger asChild let:builder>
            <Button builders={[builder]} type="text">
              {#if advancedOpened}
                <CaretDownFilledIcon size="12px" />
              {:else}
                <CaretRightFilledIcon size="12px" />
              {/if}
              <span class="text-sm">Advanced options</span>
            </Button>
          </Collapsible.Trigger>
          <Collapsible.Content class="flex flex-col gap-y-2">
            <Input
              bind:value={$form.branch}
              errors={$errors?.branch}
              id="branch"
              label="Branch"
              capitalizeLabel={false}
            />

            <Input
              bind:value={$form.subpath}
              errors={$errors?.subpath}
              id="subpath"
              label="Subpath"
              capitalizeLabel={false}
              optional
            />
          </Collapsible.Content>
        </Collapsible.Root>
      {/if}

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
      {#if isNewRepoType}
        <Button
          form={FORM_ID}
          submitForm
          type="primary"
          loading={disableSubmit}
          disabled={disableSubmit}
        >
          Create and push changes
        </Button>
      {:else}
        <Button
          type="primary"
          loading={disableSubmit}
          disabled={disableSubmit}
          onClick={() => (showOverwriteConfirmation = true)}
        >
          Pull changes
        </Button>
      {/if}
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<GithubOverwriteConfirmDialog
  bind:open={showOverwriteConfirmation}
  loading={isPending}
  error={parsedError?.message}
  githubRemote={$form.repo}
  subpath={$form.subpath}
  onConfirm={() => void submit()}
/>
