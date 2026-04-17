<script module lang="ts">
  export const CreateProjectFormId = "create-project-form";
</script>

<script lang="ts">
  import {
    createAdminServiceCreateProject,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import type { AxiosError } from "axios";
  import { sanitizeSlug } from "@rilldata/web-common/lib/string-utils.ts";
  import { createProjectAndInitEmpty } from "@rilldata/web-admin/features/projects/create-project-and-init-empty.ts";
  import {
    type DeployError,
    getPrettyDeployError,
  } from "@rilldata/web-common/features/project/deploy/deploy-errors.ts";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";

  const {
    organization,
    defaultName = "new_project",
    onCreate,
    onDeployError,
  }: {
    organization: string;
    defaultName?: string;
    onCreate: (frontendUrl: string) => void;
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
      displayName: string(),
    }),
  );

  const createProjectMutation = createAdminServiceCreateProject();

  let billingIssueMessage = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let isOrgOnTrial = $derived(!!$billingIssueMessage.data?.trial);

  const { form, tainted, errors, enhance, submit, submitting } = superForm(
    defaults({ name: defaultName, displayName: defaultName }, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const frontendUrl = await createProjectAndInitEmpty(
          $createProjectMutation,
          organization,
          form.data.name,
          form.data.displayName,
        );
        if (frontendUrl) {
          onCreate(frontendUrl);
        }
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
    },
  );

  // As a convenience, we auto generate a project name based on the display name.
  // But the moment the project name is directly changed,
  // we should stop doing this since the user probably changed it directly.
  function updateName(displayName: string) {
    if ($tainted?.name) return;
    form.update(
      ($form) => {
        $form.name = sanitizeSlug(displayName);
        return $form;
      },
      { taint: false },
    );
  }
  let displayName = $derived($form.displayName);
  $effect(() => updateName(displayName));
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
    bind:value={$form.displayName}
    errors={$errors?.displayName}
    id="displayName"
    label="Name"
    textClass="text-sm"
    alwaysShowError
    claimFocusOnMount
    width="400px"
    size="xl"
    capitalizeLabel={false}
  />
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label="URL"
    textClass="text-sm"
    alwaysShowError
    width="400px"
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
