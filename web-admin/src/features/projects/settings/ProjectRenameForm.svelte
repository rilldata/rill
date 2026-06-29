<script module lang="ts">
  export const ProjectRenameFormId = "project-update-form";
</script>

<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseUpdateProjectError } from "@rilldata/web-admin/features/projects/settings/errors";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { sanitizeSlug } from "@rilldata/web-common/lib/string-utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    organization,
    project,
    loading = $bindable(false),
    changed = $bindable(false),
    onRename,
  }: {
    organization: string;
    project: string;
    loading: boolean;
    changed: boolean;
    onRename?: (newProject: string) => void;
  } = $props();

  const initialValues: {
    name: string;
  } = {
    name: "",
  };
  const schema = yup(
    object({
      name: string().required(),
    }),
  );

  const updateProjectMutation = createAdminServiceUpdateProject();

  const { form, tainted, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const newProject = sanitizeSlug(values.name);

        try {
          await $updateProjectMutation.mutateAsync({
            org: organization,
            project,
            data: {
              newName: newProject,
            },
          });

          await queryClient.invalidateQueries({
            queryKey:
              getAdminServiceListProjectsForOrganizationQueryKey(organization),
          });
        } catch (err) {
          const parsedErr = parseUpdateProjectError(
            err as AxiosError<RpcStatus>,
          );
          if (parsedErr.duplicateProject) {
            form.errors.name = [m.settings_name_already_taken({ name: newProject })];
          }
          return;
        }

        if (project !== newProject) {
          queryClient.removeQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          });
        } else {
          void queryClient.refetchQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          });
        }
        eventBus.emit("notification", {
          message: m.settings_updated_project_notification(),
        });
        setTimeout(() => onRename?.(newProject));
      },
      resetForm: false,
    },
  );

  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );

  // Only sync server data to form on initial load or when form hasn't been modified
  $effect(() => {
    if (!$projectResp.data?.project || $tainted?.["name"]) return;
    form.update(
      (f) => {
        f.name = $projectResp.data.project.name ?? "";
        return f;
      },
      {
        taint: false,
      },
    );
  });

  $effect(() => {
    changed = $projectResp.data?.project?.name !== $form.name;
  });
  $effect(() => {
    loading = $submitting || $projectResp.isPending;
  });

  let error = $derived(
    parseUpdateProjectError(
      $updateProjectMutation.error as unknown as AxiosError<RpcStatus>,
    ),
  );
</script>

<form
  id={ProjectRenameFormId}
  onsubmit={(e) => {
    e.preventDefault();
    submit(e);
  }}
  class="update-project-form"
  use:enhance
>
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label={m.settings_name_label()}
    description={m.settings_project_url_description({ org: organization, slug: sanitizeSlug($form.name) })}
    textClass="text-sm"
    alwaysShowError
    additionalClass="max-w-[520px]"
  />

  {#if error?.message}
    <div class="text-red-500 text-sm py-px">
      {error.message}
    </div>
  {/if}
</form>

<style lang="postcss">
  .update-project-form {
    @apply flex flex-col gap-y-5 w-full;
  }
</style>
