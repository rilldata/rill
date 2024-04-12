<script lang="ts">
  import { goto } from "$app/navigation";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { runtime } from "../../runtime-client/runtime-store";
  import { renameFileArtifact } from "./actions";
  import {
    getFileAPIPathFromNameAndType,
    getLabel,
    getRouteFromName,
  } from "./entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "./name-utils";

  export let closeModal: () => void;
  export let entityType: EntityType;
  export let currentAssetName: string;

  let error: string;

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const { form, errors, handleSubmit } = createForm({
    initialValues: {
      newName: currentAssetName,
    },
    validationSchema: yup.object({
      newName: yup
        .string()
        .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
        .required("Enter a name!")
        .notOneOf([currentAssetName], `That's the current name!`),
    }),
    onSubmit: async (values) => {
      if (
        isDuplicateName(
          values?.newName,
          currentAssetName,
          $allNamesQuery?.data ?? [],
        )
      ) {
        error = `Name ${values.newName} is already in use`;
        return;
      }
      try {
        await renameFileArtifact(
          runtimeInstanceId,
          getFileAPIPathFromNameAndType(currentAssetName, entityType),
          getFileAPIPathFromNameAndType(values.newName, entityType),
          entityType,
        );
        const edit = entityType === EntityType.MetricsDefinition ? "/edit" : "";
        await goto(getRouteFromName(values.newName, entityType) + edit, {
          replaceState: true,
        });
        closeModal();
      } catch (err) {
        error = err.response.data.message;
      }
    },
  });

  $: entityLabel = getLabel(entityType);
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
          id="{entityLabel}-name"
          label="{entityLabel} name"
        />
      </div>
    </form>
  </div>
  <svelte:fragment slot="primary-action-body">Change Name</svelte:fragment>
</Dialog>
