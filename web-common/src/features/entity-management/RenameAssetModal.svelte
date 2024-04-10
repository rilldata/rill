<script lang="ts">
  import { goto } from "$app/navigation";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";
  import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { extractFileExtension } from "@rilldata/web-common/features/sources/extract-file-name";
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

  let error: string;

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const [folder, assetName] = splitFolderAndName(filePath);

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
        isDuplicateName(values?.newName, assetName, $allNamesQuery?.data ?? [])
      ) {
        error = `Name ${values.newName} is already in use`;
        return;
      }
      try {
        const newPath = (folder ? `${folder}/` : "") + values.newName;
        await renameFileArtifact(runtimeInstanceId, filePath, newPath);
        goto(`/files/${newPath}`, {
          replaceState: true,
        });
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
          id="entity-name"
          label="entity name"
        />
      </div>
    </form>
  </div>
  <svelte:fragment slot="primary-action-body">Change Name</svelte:fragment>
</Dialog>
