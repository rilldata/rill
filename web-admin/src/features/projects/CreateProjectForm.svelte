<script module lang="ts">
  export const CreateProjectFormId = "create-project-form";
</script>

<script lang="ts">
  import {
    createAdminServiceCreateDeployment,
    createAdminServiceCreateManagedGitRepo,
    createAdminServiceCreateProject,
    getAdminServiceListDeploymentsQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import type { AxiosError } from "axios";
  import {
    type DeployError,
    getPrettyDeployError,
  } from "@rilldata/web-common/features/project/deploy/deploy-errors.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { CreateProjectBranchName } from "@rilldata/web-admin/features/projects/publish-project.ts";

  const {
    organization,
    defaultName = "new_project",
    onCreate,
    onDeployError,
  }: {
    organization: string;
    defaultName?: string;
    onCreate: (projectName: string, frontendUrl: string) => void;
    onDeployError?: (deployError: DeployError) => void;
  } = $props();

  const schema = yup(
    object({
      name: string()
        .required("Name is required")
        .matches(
          /^[a-zA-Z0-9][a-zA-Z0-9_-]*$/,
          "Name must start with a letter or number and may only contain letters, numbers, hyphens, and underscores",
        )
        .min(1, "Name must be at least 1 character")
        .max(40, "Name must be at most 40 characters"),
    }),
  );

  const createManagedGitRepo = createAdminServiceCreateManagedGitRepo();
  const createProjectMutation = createAdminServiceCreateProject();
  const createDeployMutation = createAdminServiceCreateDeployment();

  let billingIssueMessage = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let isOrgOnTrial = $derived(!!$billingIssueMessage.data?.trial);

  let createdGitRepo = $state("");

  let formInstance = $derived(
    superForm(defaults({ name: defaultName }, schema), {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const project = form.data.name;

        // Step 1: Create the git repo.

        // As an optimization, we only create the git repo once.
        // Note that this is not really persisted across page reloads.
        // We dont really need it since the orphaned repos are deleted eventually.
        if (!createdGitRepo) {
          const createManagedGitRepoResult =
            await $createManagedGitRepo.mutateAsync({
              org: organization,
              data: { name: project },
            });
          createdGitRepo = createManagedGitRepoResult.remote ?? "";
        }

        // Step 2: Create the project.
        const resp = await $createProjectMutation.mutateAsync({
          org: organization,
          data: {
            project,
            gitRemote: createdGitRepo,
            prodSlots: "4",
            skipDeploy: true,
          },
        });
        void queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListProjectsForOrganizationQueryKey(organization),
        });

        // Step 3: Create the editable dev deployment.
        // TODO: Handle deployment creation failure. Project would be created leading to possible duplicate project error on retry.
        await $createDeployMutation.mutateAsync({
          org: organization,
          project,
          data: {
            environment: "dev",
            branch: CreateProjectBranchName,
            editable: true,
          },
        });
        void queryClient.invalidateQueries({
          queryKey: getAdminServiceListDeploymentsQueryKey(
            organization,
            project,
          ),
        });

        onCreate(project, resp.project?.frontendUrl ?? "/");
      },
      onError({ result }) {
        const error =
          (result.error as AxiosError<RpcStatus>)?.response?.data?.message ??
          result.error.message;
        if (!error) return;
        // Mapping for backend error to a more user friendly UI error message.
        if (error.includes("a project with that name already exists")) {
          $errors["name"] = [
            `Project name '${$form.name}' is already taken. Please try a different name.`,
          ];
        } else {
          const deployError = getPrettyDeployError(
            new Error(error),
            isOrgOnTrial,
          );
          if (deployError) onDeployError?.(deployError);
          $errors["name"] = [error];
        }
      },
    }),
  );

  let { form, errors, enhance, submit, submitting } = $derived(formInstance);
</script>

<form
  id={CreateProjectFormId}
  onsubmit={(e) => {
    e.preventDefault();
    submit(e);
  }}
  use:enhance
  class="flex flex-col gap-y-4"
>
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label="Name"
    textClass="text-sm"
    alwaysShowError
    width="500px"
    size="xl"
    textInputPrefix="https://ui.rilldata.com/{organization}/"
  />
  <div class="w-full flex justify-end">
    <Button
      type="primary"
      submitForm
      loading={$submitting}
      disabled={$submitting}
      onClick={submit}
    >
      Create project
    </Button>
  </div>
</form>
