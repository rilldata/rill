<script lang="ts">
  import { goto, invalidate } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { uploadTableFiles } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { createRuntimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { yup } from "sveltekit-superforms/adapters";
  import { defaults, superForm, filesProxy } from "sveltekit-superforms";
  import { object, array, mixed } from "yup";
  import ShadcnInput from "@rilldata/web-common/components/forms/ShadcnInput.svelte";
  import { UploadFileSizeLimitInBytes } from "@rilldata/web-common/features/entity-management/file-selectors.ts";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";
  import {
    PossibleFileExtensions,
    PossibleZipExtensions,
  } from "@rilldata/web-common/features/sources/modal/possible-file-extensions.ts";
  import { onMount } from "svelte";

  export let initFiles: File[] = [];
  export let onClose: () => void = () => {};
  export let onBack: () => void = () => {};
  export let showBack = false;

  const FORM_ID = "upload-files-form";
  const schema = yup(
    object({
      files: array().of(
        mixed().test(
          "fileSize",
          "Local files over 100 MB can’t be deployed to Rill Cloud. Please upload to S3 or another external storage first if you want to deploy to Rill Could.",
          (value) => {
            // Check if value exists and its size is within the limit
            return value && value.size <= UploadFileSizeLimitInBytes;
          },
        ),
      ),
    }),
  );

  const { form, submit, enhance, errors } = superForm(
    defaults({ files: initFiles }, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;

        const values = form.data;
        await handleUpload(values.files);
      },
      validationMethod: "oninput",
    },
  );

  const files = filesProxy(form, "files");
  $: filesError = Object.values($errors.files ?? {})[0] ?? "";

  $: ({ instanceId } = $runtime);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(files, instanceId, false);
    const initialized = await isProjectInitialized(instanceId);
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        // If project is uninitialized, initialize an empty project
        if (!initialized) {
          $unpackEmptyProject.mutate({
            instanceId,
            data: {
              displayName: EMPTY_PROJECT_TITLE,
              olap: "duckdb", // Explicitly set DuckDB as OLAP for local file uploads
            },
          });

          // Race condition: invalidate("init") must be called before we navigate to
          // `/files/${newFilePath}`. invalidate("init") is also called in the
          // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
          await invalidate("init");
        }

        const yaml = compileLocalFileSourceYAML(filePath);
        await createSource(instanceId, tableName, yaml);
        const newFilePath = getFilePathFromNameAndType(
          tableName,
          EntityType.Table,
        );
        await goto(`/files${newFilePath}`);
      } catch (err) {
        console.error(err);
      }

      overlay.set(null);
      onClose();
    }
  }

  onMount(() => {
    if (initFiles?.length) {
      submit();
    }
  });
</script>

<form
  method="POST"
  enctype="multipart/form-data"
  id={FORM_ID}
  use:enhance
  on:submit|preventDefault={submit}
>
  <div class="grid place-items-center h-44">
    <div class="flex flex-col w-96">
      <ShadcnInput
        type="file"
        bind:files={$files}
        multiple
        accept={[...PossibleFileExtensions, ...PossibleZipExtensions].join(",")}
        class="h-8 w-full"
      />
      {#if filesError}
        <div class="text-red-600 text-xs py-px mt-0.5">
          <div>{filesError}</div>
          <ul class="flex flex-col list-disc ml-5 gap-y-1">
            {#each $files as file, i (file.name)}
              {#if $errors.files?.[i]}
                {@const formattedSize = formatMemorySize(
                  Number(file.size || 0),
                )}
                <li>{file.name} ({formattedSize})</li>
              {/if}
            {/each}
          </ul>
        </div>
      {/if}
    </div>
  </div>
  <div class="flex p-6 gap-x-1">
    <div class="grow" />
    {#if showBack}
      <Button onClick={onBack} type="secondary">Back</Button>
    {/if}
    <Button form={FORM_ID} submitForm type="primary">Upload</Button>
  </div>
</form>
