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
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
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

  let activeTab = "new";
  let showOverwriteConfirmation = false;

  const githubUserOrgs = getGithubUserOrgs();
  $: githubUserRepos = getGithubUserRepos(activeTab === "existing");

  // Keep $form.type in sync with the active tab so yup conditional
  // validation continues to work unchanged.
  $: $form.type = activeTab === "new" ? "create" : "pull";

  const initialValues: {
    type: string;
    repo: string;
    org: string;
    name: string;
    branch: string;
    subpath: string;
  } = {
    type: "create",
    repo: "",
    org: "",
    name: project,
    branch: "",
    subpath: "",
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
        if (!form.valid) return;
        const values = form.data;

        if (activeTab === "new") {
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
              primaryBranch: values.branch,
              subpath: values.subpath,
            },
          });
        }
      },
    },
  );

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
      activeTab = "new";
    }
  }}
>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        type="primary"
        class="w-fit mt-1"
        loading={$userStatus.isFetching}
        onClick={() => void githubAccessManager.ensureGithubAccess()}
      >
        <Github className="w-5 h-5 flex-shrink-0" />
        {m.github_connect_to_github()}
      </Button>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <div class="flex flex-col gap-y-1">
          <Dialog.Title>{m.github_connect_to_github()}</Dialog.Title>
          <Dialog.Description>
            {m.github_connect_project_description()}
          </Dialog.Description>
        </div>
      </div>
    </Dialog.Header>

    <Tabs
      value={activeTab}
      onValueChange={(value) => {
        if (value) activeTab = value;
        resetMutations();
        reset();
      }}
      class="mt-1"
    >
      <UnderlineTabsList>
        <UnderlineTabsTrigger value="new"
          >{m.github_create_repository()}</UnderlineTabsTrigger
        >
        <UnderlineTabsTrigger value="existing">
          {m.github_existing_repository()}
        </UnderlineTabsTrigger>
      </UnderlineTabsList>

      <form
        id={FORM_ID}
        onsubmit={(e) => {
          e.preventDefault();
          submit(e);
        }}
        use:enhance
        class="flex flex-col gap-y-3 pt-4"
      >
        {#if activeTab === "new"}
          <Select
            bind:value={$form.org}
            id="org"
            label={m.github_organization()}
            placeholder={m.github_select_organization()}
            options={$githubUserOrgs.data ?? []}
            optionsLoading={$githubUserOrgs.isFetching}
            sameWidth
            enableSearch
            onAddNew={() => githubAccessManager.reselectOrgOrRepos(false)}
            addNewLabel={m.github_connect_other_orgs()}
          />

          <Input
            bind:value={$form.name}
            errors={$errors?.name}
            id="name"
            label={m.github_repository_name()}
            capitalizeLabel={false}
          />
        {:else}
          <Select
            bind:value={$form.repo}
            id="repo"
            label={m.github_repository()}
            placeholder={m.github_select_repository()}
            sameWidth
            options={$githubUserRepos.data?.repoOptions ?? []}
            optionsLoading={$githubUserRepos.isFetching}
            enableSearch
            onChange={(newRepo) => onSelectedRepoChange(newRepo)}
            onAddNew={() => githubAccessManager.reselectOrgOrRepos(true)}
            addNewLabel={m.github_connect_other_repos()}
          />

          <Input
            bind:value={$form.branch}
            errors={$errors?.branch}
            id="branch"
            label={m.github_branch()}
            placeholder="main"
            capitalizeLabel={false}
          />

          <Input
            bind:value={$form.subpath}
            errors={$errors?.subpath}
            id="subpath"
            label={m.github_subpath()}
            placeholder="/"
            capitalizeLabel={false}
            optional
          />
        {/if}

        {#if parsedError?.message}
          <div class="text-red-500 text-sm py-px">
            {parsedError.message}
          </div>
        {/if}

        {#if $githubConnectionFailed}
          <div class="text-red-500 text-sm py-px">
            <div>{m.github_connection_failed()}</div>
            <Button
              type="secondary"
              onClick={() => githubAccessManager.ensureGithubAccess()}
            >
              {m.github_reconnect()}
            </Button>
          </div>
        {/if}
      </form>
    </Tabs>

    <Dialog.Footer>
      <Button
        onClick={() => {
          open = false;
          reset();
          activeTab = "new";
        }}
        type="secondary"
      >
        {m.common_cancel()}
      </Button>
      {#if activeTab === "new"}
        <Button
          form={FORM_ID}
          submitForm
          type="primary"
          loading={disableSubmit}
          disabled={disableSubmit}
        >
          {m.github_create_and_push()}
        </Button>
      {:else}
        <Button
          type="primary"
          loading={disableSubmit}
          disabled={disableSubmit}
          onClick={() => (showOverwriteConfirmation = true)}
        >
          {m.github_pull_changes()}
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
