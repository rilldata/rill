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

  const {
    organization,
    defaultName = "new_project",
    onCreate,
  }: {
    organization: string;
    defaultName?: string;
    onCreate: (frontendUrl: string) => void;
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

  const createProjectMutation = createAdminServiceCreateProject();

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults({ name: defaultName }, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const resp = await $createProjectMutation.mutateAsync({
          org: organization,
          data: {
            project: form.data.name,
            generateManagedGit: true,
            prodSlots: "4",
          },
        });
        const frontendUrl = resp.project?.frontendUrl;
        if (!frontendUrl) return;
        onCreate(frontendUrl);
      },
      onError({ result }) {
        const error =
          (result.error as AxiosError<RpcStatus>)?.response.data.message ??
          result.error.message;
        if (!error) return;
        // Mapping for backend error to a more user friendly UI error message.
        if (error.includes("a project with that name already exists")) {
          $errors["name"] = [
            `Project name '${$form.name}' is already taken. Please try a different name.`,
          ];
        } else {
          $errors["name"] = [error];
        }
      },
    },
  );
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
    label="Project name"
    textClass="text-sm"
    alwaysShowError
    claimFocusOnMount
    width="400px"
    size="xl"
    capitalizeLabel={false}
  />
  <Button
    type="primary"
    wide
    submitForm
    loading={$submitting}
    disabled={$submitting}
    onClick={submit}
  >
    Create project
  </Button>
</form>
