<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";
  import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import {
    useAllFileNames,
    useDirectoryNamesInDirectory,
  } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "../../runtime-client/runtime-store";
  import { renameFileArtifact } from "./actions";
  import { removeLeadingSlash } from "./entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "./name-utils";
  import { superForm, defaults } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let closeModal: () => void;
  export let filePath: string;
  export let isDir: boolean;

  let error: string;

  const [folder, assetName] = splitFolderAndName(filePath);

  const validationSchema = object({
    newName: string()
      .required("Enter a name!")
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE),
  });

  const initialValues = {
    newName: assetName,
  };

  const {
    form: superform,
    enhance,
    submit,
    errors,
  } = superForm(defaults(initialValues, yup(validationSchema)), {
    SPA: true,
    validators: yup(validationSchema),
    async onUpdate({ form }) {
      if (!form.valid) return;

      const values = form.data;

      if (values.newName === assetName) {
        closeModal();
        return;
      }

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
          if (
            $page.url.pathname.startsWith(
              `/files/${removeLeadingSlash(filePath)}`,
            )
          ) {
            // if the file focused has the dir then replace the dir path to the new one
            void goto(
              $page.url.pathname.replace(
                `/files/${removeLeadingSlash(filePath)}`,
                `/files/${removeLeadingSlash(newPath)}`,
              ),
            );
          }
        } else {
          await goto(`/files/${removeLeadingSlash(newPath)}`, {
            replaceState: true,
          });
        }
        closeModal();
      } catch (err) {
        error = err.response.data.message;
      }
    },
  });

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllFileNames(queryClient, runtimeInstanceId);

  $: existingDirectories = useDirectoryNamesInDirectory(
    runtimeInstanceId,
    folder,
  );
</script>

<Dialog
  compact
  on:cancel={closeModal}
  on:click-outside={closeModal}
  on:primary-action={submit}
  showCancel
  size="sm"
>
  <svelte:fragment slot="title">Rename</svelte:fragment>
  <div slot="body">
    {#if error}
      <SubmissionError message={error} />
    {/if}
    <form autocomplete="off" on:submit|preventDefault={submit} use:enhance>
      <div class="py-2">
        <Input
          bind:value={$superform.newName}
          claimFocusOnMount
          alwaysShowError
          errors={$errors.newName?.[0]}
          id={isDir ? "folder-name" : "file-name"}
          label={isDir ? "Folder name" : "File name"}
        />
      </div>
    </form>
  </div>
  <svelte:fragment slot="primary-action-body">Change Name</svelte:fragment>
</Dialog>
