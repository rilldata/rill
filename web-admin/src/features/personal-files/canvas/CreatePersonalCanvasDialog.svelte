<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
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
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { setCanvasMode } from "@rilldata/web-admin/features/personal-files/canvas/mode-utils.ts";

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
    }),
  );

  const schema = yup(
    object({
      name: string().trim().required("Name is required"),
      mode: string(),
      copySource: string().when("mode", {
        is: "copy",
        then: (schema) => schema.required("Copy source is required"),
        otherwise: (schema) => schema.notRequired(),
      }),
    }),
  );
  const initialValues: {
    name: string;
    mode: "blank" | "copy";
    copySource: string;
  } = {
    name: "",
    mode: "blank",
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
        if (values.mode === "blank") {
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

  function updateMode(newMode: "blank" | "copy") {
    form.update((f) => ({ ...f, mode: newMode }));
  }

  let error = $derived(
    $createFileMutation.error?.message ?? $errors["copySource"]?.[0],
  );
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <Button {...props} type="primary">Create dashboard</Button>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Create personal dashboard</Dialog.Title>
      <Dialog.Description>
        Personal dashboards are only visible to you. They live alongside the
        project but never sync to git.
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
        label="Display name"
        placeholder="e.g. My revenue dashboard"
      />

      <Tabs value={$form.mode} onValueChange={updateMode} class="mt-1">
        {#if personalCanvasOptions.length > 0}
          <UnderlineTabsList>
            <UnderlineTabsTrigger value="blank">
              Blank dashboard
            </UnderlineTabsTrigger>
            <UnderlineTabsTrigger value="copy">
              Copy from an existing dashboard
            </UnderlineTabsTrigger>
          </UnderlineTabsList>
        {/if}

        {#if $form.mode === "copy"}
          <Select
            bind:value={$form.copySource}
            id="source"
            placeholder="Select a dashboard..."
            options={personalCanvasOptions}
            optionsLoading={$personalCanvasesQuery.isPending}
            sameWidth
            enableSearch
          />
        {/if}
      </Tabs>

      {#if error}
        <p class="text-destructive text-sm">{error}</p>
      {/if}
    </form>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        onClick={submit}
        loading={$submitting}
        loadingCopy="Creating..."
        disabled={$submitting}
      >
        Create
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
