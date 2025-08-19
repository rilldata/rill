<script lang="ts">
  import {
    createAdminServiceConnectProjectToGithub,
    createAdminServiceUpdateProject,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils.ts";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData.ts";
  import GithubOverwriteConfirmDialog from "@rilldata/web-admin/features/projects/github/GithubOverwriteConfirmDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CaretDownFilledIcon from "@rilldata/web-common/components/icons/CaretDownFilledIcon.svelte";
  import CaretRightFilledIcon from "@rilldata/web-common/components/icons/CaretRightFilledIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { SelectSeparator } from "@rilldata/web-common/components/select";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import { derived } from "svelte/store";
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
  $: mutationStatus = derived(
    [connectProjectToGithub, updateProject],
    ([$connect, $update]) => {
      return {
        error: $connect.error ?? $update.error,
        isPending: $connect.isPending || $update.isPending,
      };
    },
  );
  $: ({ error, isPending } = $mutationStatus);

  let isNewRepoType = false;
  let isPushRepoType = false;
  let remote = "";
  let advancedOpened = false;
  let showOverwriteConfirmation = false;

  type GithubSelectionType = "new" | "pull" | "push";
  const GithubSelectionTypeOptions = [
    {
      label: "Push project to a new Github repo",
      buttonLabel: "Create and push changes",
      value: "new",
    },
    {
      label: "Pull changes from existing Github repo",
      buttonLabel: "Pull changes",
      value: "pull",
    },
    {
      label: "Push changes to existing Github repo",
      buttonLabel: "Overwrite",
      value: "push",
    },
  ];
  let selectedTypeOption = GithubSelectionTypeOptions[0];

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
    branch: "main",
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
      subpath: string(),
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

        if (isNewRepoType || isPushRepoType) {
          await $connectProjectToGithub.mutateAsync({
            project: project,
            organization: organization,
            data: {
              remote,
              branch: values.branch,
              subpath: values.subpath,
              create: isNewRepoType,
              // We always show the confirmation so we can force push.
              force: true,
            },
          });
        } else {
          await $updateProject.mutateAsync({
            name: project,
            organizationName: organization,
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
  $: remote = isNewRepoType
    ? `https://github.com/${$form.org}/${$form.name}.git`
    : $form.repo;
  $: selectedTypeOption = GithubSelectionTypeOptions.find(
    (o) => o.value === $form.type,
  )!;

  function onConnectToGithub() {
    void githubData.startRepoSelection();
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: false, // TODO
      },
    );
  }

  function onSelectedRepoChange(newRemote: string) {
    const repo = $userRepos.data?.repos?.find((r) => r.remote === newRemote);
    if (!repo?.defaultBranch) return;

    $form.branch = repo.defaultBranch;
  }
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
          <Select
            bind:value={$form.org}
            id="org"
            label="Repository org"
            options={orgSelections}
            sameWidth
            enableSearch
          >
            <div slot="additional-dropdown-content">
              <SelectSeparator />
              <button
                on:click={() => githubData.reselectRepos()}
                class="w-full cursor-pointer select-none rounded-sm py-1.5 px-2 text-left hover:bg-accent"
                type="button"
              >
                + Connect other orgs
              </button>
            </div>
          </Select>

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
            options={repoSelections}
            enableSearch
            onChange={(newRepo) => onSelectedRepoChange(newRepo)}
          >
            <div slot="additional-dropdown-content">
              <SelectSeparator />
              <button
                on:click={() => githubData.reselectRepos()}
                class="w-full cursor-pointer select-none rounded-sm py-1.5 px-2 text-left hover:bg-accent"
                type="button"
              >
                + Connect other repos
              </button>
            </div>
          </Select>
        {/if}

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
          <Collapsible.Content class="ml-6 flex flex-col gap-y-2">
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

        {#if isPushRepoType}
          <div class="text-sm text-muted-foreground">
            Project will replace contents of this repo.
          </div>
        {/if}

        {#if error?.message}
          <div class="text-red-500 text-sm py-px">
            {error.message}
          </div>
        {/if}
      </form>
    </Dialog.Description>
    <Dialog.Footer>
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      {#if isNewRepoType}
        <Button
          form={FORM_ID}
          submitForm
          type="primary"
          loading={isPending}
          disabled={isPending}
        >
          {selectedTypeOption.buttonLabel}
        </Button>
      {:else}
        <Button
          type="primary"
          loading={isPending}
          disabled={isPending}
          onClick={() => (showOverwriteConfirmation = true)}
        >
          {selectedTypeOption.buttonLabel}
        </Button>
      {/if}
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<GithubOverwriteConfirmDialog
  bind:open={showOverwriteConfirmation}
  loading={isPending}
  error={getRpcErrorMessage(error)}
  githubRemote={remote}
  subpath={$form.subpath}
  type={isPushRepoType ? "push" : "pull"}
  onConfirm={() => void submit()}
/>
