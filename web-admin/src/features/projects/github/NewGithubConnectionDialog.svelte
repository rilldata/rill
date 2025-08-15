<script lang="ts">
  import {
    createAdminServiceConnectProjectToGithub,
    createAdminServiceUpdateProject,
  } from "@rilldata/web-admin/client";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let organization: string;
  export let project: string;

  const FORM_ID = "github-connect-form";

  const githubData = getGithubData();
  const userStatus = githubData.userStatus;
  const userRepos = githubData.userRepos;

  $: orgSelections =
    $userStatus.data?.organizations?.map((o) => ({
      value: o,
      label: o,
    })) ?? [];
  $: repoSelections =
    $userRepos.data?.repos?.map((r) => ({
      value: r.remote,
      label: `${r.owner}/${r.name}`,
    })) ?? [];

  const connectProjectToGithub = createAdminServiceConnectProjectToGithub();
  const updateProject = createAdminServiceUpdateProject();
  let isNewRepoType = false;
  let isPushRepoType = false;

  type GithubSelectionType = "new" | "pull" | "push";
  const GithubSelectionTypeOptions = [
    {
      label: "Push project to a new Github repo",
      value: "new",
    },
    {
      label: "Pull changes from existing Github repo",
      value: "pull",
    },
    {
      label: "Push changes to existing Github repo",
      value: "push",
    },
  ];

  const initialValues: {
    type: GithubSelectionType;
    org: string;
    name: string;
    repo: string;
    branch: string;
    subpath: string;
  } = {
    type: "new",
    org: "",
    name: project, // Initialize repo name with project name
    repo: "",
    branch: "",
    subpath: "",
  };
  const schema = yup(
    object({
      type: string().required(),
      org: string().when("type", {
        is: "new",
        then: (schema) => schema.required("Org is required"),
        otherwise: (schema) => schema.notRequired(),
      }),
      name: string().when("type", {
        is: "new",
        then: (schema) => schema.required("Repo name is required"),
        otherwise: (schema) => schema.notRequired(),
      }),
      repo: string().when("type", {
        is: "new",
        then: (schema) => schema.notRequired(),
        otherwise: (schema) => schema.required("Repo is required"),
      }),
      branch: string().when("type", {
        is: "new",
        then: (schema) => schema.notRequired(),
        otherwise: (schema) => schema.required("Branch is required"),
      }),
    }),
  );

  const { form, errors, enhance, submit } = superForm(
    defaults(initialValues, schema),
    {
      id: FORM_ID,
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values = form.data;
        console.log(values);
        let remote = values.remote ?? "";
        if (isNewRepoType) {
          remote = `https://github.com/${values.org}/${values.name}.git`;
        }

        if (isNewRepoType || isPushRepoType) {
          await $connectProjectToGithub.mutateAsync({
            project: project,
            organization: organization,
            data: {
              gitRemote: remote,
              prodBranch: values.branch,
              subpath: values.subpath,
              create: isNewRepoType,
            },
          });
        } else {
          await $updateProject.mutateAsync({
            name: project,
            organization,
            data: {
              gitRemote: remote,
              prodBranch: values.branch,
              subpath: values.subpath,
            },
          });
        }
      },
    },
  );

  $: isNewRepoType = $form.type === "new";
  $: isPushRepoType = $form.type === "push";

  function onConnectToGithub() {
    void githubData.startRepoSelection();
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: false, // TODO
      },
    );
  }

  $: console.log($form, $errors);
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild let:builder>
    <Button
      builders={[builder]}
      type="primary"
      class="w-fit mt-1"
      loading={$userStatus.isFetching}
      onClick={onConnectToGithub}
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
            Connect this project to a new or existing repo.
          </Dialog.Description>
        </div>
      </div>
    </Dialog.Header>
    <Dialog.Description>
      <form
        id={FORM_ID}
        on:submit|preventDefault={submit}
        use:enhance
        class="flex flex-col gap-y-3"
      >
        <Input
          bind:value={$form.type}
          errors={$errors?.type}
          id="type"
          label="Connection type"
          capitalizeLabel={false}
          sameWidth
          options={GithubSelectionTypeOptions}
        />

        {#if isNewRepoType}
          <Input
            bind:value={$form.org}
            errors={$errors?.org}
            id="org"
            label="Repository org"
            capitalizeLabel={false}
            options={orgSelections}
            sameWidth
            enableSearch
          />

          <Input
            bind:value={$form.name}
            errors={$errors?.name}
            id="name"
            label="Repository name"
            capitalizeLabel={false}
          />
        {:else}
          <Input
            bind:value={$form.repo}
            errors={$errors?.repo}
            id="name"
            label="Repository"
            capitalizeLabel={false}
            sameWidth
            options={repoSelections}
            enableSearch
          />

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
        {/if}

        {#if isPushRepoType}
          <div class="text-sm text-muted-foreground">
            Project will replace contents of this repo.
          </div>
        {/if}
      </form>
    </Dialog.Description>
    <Dialog.Footer>
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button form={FORM_ID} submitForm type="primary">OK (TODO)</Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
