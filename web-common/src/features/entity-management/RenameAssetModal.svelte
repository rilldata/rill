<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";
  import {
    useAllFileNames,
    useDirectoryNamesInDirectory,
  } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { runtime } from "../../runtime-client/runtime-store";
  import { renameFileArtifact } from "./actions";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "./name-utils";

  export let closeModal: () => void;
  export let filePath: string;
  export let isDir: boolean;

  let error: string;
  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllFileNames(queryClient, runtimeInstanceId);

  const [folder, assetName] = splitFolderAndName(filePath);

  $: existingDirectories = useDirectoryNamesInDirectory(
    runtimeInstanceId,
    folder,
  );

  const { form, errors, handleSubmit } = createForm({
    initialValues: {
      newName: assetName,
    },
    validationSchema: yup.object({
      newName: yup
        .string()
        .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
        .required("Enter a name!")
        .notOneOf([assetName], `That's the current name!`),
    }),
    onSubmit: async (values) => {
      if (
        isDir &&
        isDuplicateName(
          values?.newName,
          assetName,
          $existingDirectories?.data ?? [],
        )
      ) {
        error = `An existing folder with name ${values.newName} already exists`;
        return;
      }
      if (
        !isDir &&
        isDuplicateName(values?.newName, assetName, $allNamesQuery?.data ?? [])
      ) {
        error = `Name ${values.newName} is already in use`;
        return;
      }
      try {
        const newPath = (folder ? `${folder}/` : "") + values.newName;
        await renameFileArtifact(runtimeInstanceId, filePath, newPath);
        if (isDir) {
          if ($page.url.pathname.startsWith(`/files/${filePath}`)) {
            // if the file focused has the dir then replace the dir path to the new one
            void goto(
              $page.url.pathname.replace(
                `/files/${filePath}`,
                `/files/${newPath}`,
              ),
            );
          }
        } else {
          await goto(`/files${newPath}`, {
            replaceState: true,
          });
        }
        closeModal();
      } catch (err) {
        error = err.response.data.message;
      }
    },
  });
</script>

<Dialog
  compact
  disabled={$form["newName"] === ""}
  on:cancel={closeModal}
  on:click-outside={closeModal}
  on:primary-action={handleSubmit}
  showCancel
  size="sm"
>
  <svelte:fragment slot="title">Rename</svelte:fragment>
  <div slot="body">
    {#if error}
      <SubmissionError message={error} />
    {/if}
    <form autocomplete="off" on:submit|preventDefault={handleSubmit}>
      <div class="py-2">
        <Input
          bind:value={$form["newName"]}
          claimFocusOnMount
          error={$errors["newName"]}
          id={isDir ? "folder-name" : "file-name"}
          label={isDir ? "Folder name" : "File name"}
        />
      </div>
    </form>
  </div>
  <svelte:fragment slot="primary-action-body">Change Name</svelte:fragment>
</Dialog>
