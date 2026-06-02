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
  } from "@rilldata/web-admin/client";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/virtual-file-editor/selectors.ts";

  let {
    open = $bindable(false),
    org,
    project,
  }: {
    open: boolean;
    org: string;
    project: string;
  } = $props();

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
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Create personal canvas</Dialog.Title>
      <Dialog.Description>
        Personal canvases are only visible to you. They live alongside the
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
              Blank canvas
            </UnderlineTabsTrigger>
            <UnderlineTabsTrigger value="copy">
              Copy from an existing canvas
            </UnderlineTabsTrigger>
          </UnderlineTabsList>
        {/if}

        {#if $form.mode === "copy"}
          <Select
            bind:value={$form.copySource}
            id="source"
            placeholder="Select a canvas..."
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
