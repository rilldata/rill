<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { yup } from "sveltekit-superforms/adapters";
  import { boolean, object, string } from "yup";
  import { defaults, superForm } from "sveltekit-superforms";
  import { generateBlobForNewResourceFile } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    adminServiceGetPersonalFile,
    createAdminServiceCreatePersonalFile,
    getAdminServiceGetPersonalFileQueryKey,
    getAdminServiceListPersonalFilesQueryKey,
  } from "@rilldata/web-admin/client";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { setCanvasMode } from "@rilldata/web-admin/features/personal-files/canvas/mode-utils.ts";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    org,
    project,
  }: {
    org: string;
    project: string;
  } = $props();

  let open = $state(false);

  const runtimeClient = useRuntimeClient();

  const createFileMutation = createAdminServiceCreatePersonalFile();
  let personalCanvasesQuery = $derived(
    getPersonalFilteredResources(
      runtimeClient,
      org,
      project,
      ResourceKind.Canvas,
    ),
  );
  let personalCanvasOptions = $derived(
    $personalCanvasesQuery.data?.map((r) => {
      const name = r.meta?.name?.name ?? "";
      const displayName = r.canvas?.state?.validSpec?.displayName ?? name;
      return { value: name, label: displayName };
    }) ?? [],
  );

  const schema = yup(
    object({
      name: string().trim().required(m.personal_files_create_name_required()),
      copy: boolean(),
      copySource: string().when("copy", {
        is: true,
        then: (schema) =>
          schema.required(m.personal_files_create_copy_source_required()),
        otherwise: (schema) => schema.notRequired(),
      }),
    }),
  );
  const initialValues: {
    name: string;
    copy: boolean;
    copySource: string;
  } = {
    name: "",
    copy: false,
    copySource: "",
  };

  const { form, formId, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values = form.data;

        let yaml = "";
        if (!values.copy) {
          yaml = generateBlobForNewResourceFile(ResourceKind.Canvas);
        } else {
          const sourceFile = await queryClient.fetchQuery({
            queryKey: getAdminServiceGetPersonalFileQueryKey(
              org,
              project,
              values.copySource,
            ),
            queryFn: () =>
              adminServiceGetPersonalFile(org, project, values.copySource),
          });
          yaml = sourceFile.yaml;
        }

        const createResp = await $createFileMutation.mutateAsync({
          org,
          project,
          data: {
            displayName: values.name,
            kind: ResourceKind.Canvas,
            yaml,
          },
        });

        // Invalidate resources and personal files queries
        await queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            runtimeClient.instanceId,
            {},
          ),
        });
        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListPersonalFilesQueryKey(org, project),
        });

        const name = createResp.name ?? values.name;
        setCanvasMode(org, project, name, "edit");
        await goto(
          `/${org}/${project}/-/personal/${createResp.name ?? values.name}`,
        );
      },
    },
  );

  let error = $derived(
    $createFileMutation.error?.message ?? $errors["copySource"]?.[0],
  );
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <Button {...props} type="primary">
        {m.personal_files_create_trigger()}
      </Button>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content class="top-[30%] translate-y-0">
    <Dialog.Header>
      <Dialog.Title>{m.personal_files_create_title()}</Dialog.Title>
      <Dialog.Description>
        {m.personal_files_create_description()}
      </Dialog.Description>
    </Dialog.Header>

    <form
      id={$formId}
      onsubmit={(e) => {
        e.preventDefault();
        submit(e);
      }}
      use:enhance
      class="flex flex-col gap-y-3 pt-4"
    >
      <Input
        bind:value={$form.name}
        id="name"
        label={m.personal_files_create_name_label()}
        placeholder={m.personal_files_create_name_placeholder()}
      />

      {#if personalCanvasOptions.length}
        <div class="flex items-center space-x-2">
          <Switch
            bind:checked={$form["copy"]}
            id="copy"
            label={m.personal_files_create_copy_label()}
          />
          <Label class="font-normal flex gap-x-1 items-center" for="copy">
            {m.personal_files_create_copy_hint()}
          </Label>
        </div>

        {#if $form.copy}
          <Select
            bind:value={$form.copySource}
            id="source"
            placeholder={m.personal_files_create_copy_source_placeholder()}
            options={personalCanvasOptions}
            optionsLoading={$personalCanvasesQuery.isPending}
            sameWidth
            enableSearch
          />
        {/if}
      {/if}

      {#if error}
        <p class="text-destructive text-sm">{error}</p>
      {/if}
    </form>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>
        {m.personal_files_create_cancel()}
      </Button>
      <Button
        type="primary"
        onClick={submit}
        loading={$submitting}
        loadingCopy={m.personal_files_create_submitting()}
        disabled={$submitting}
      >
        {m.personal_files_create_submit()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
