<script lang="ts">
  import { goto } from "$app/navigation";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { Dialog } from "@rilldata/web-common/components/modal/index";
  import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useRuntimeServiceRenameFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { renameFileArtifact } from "./actions";
  import { getLabel, getRouteFromName } from "./entity-mappers";
  import { isDuplicateName } from "./name-utils";
  import { useAllNames } from "./selectors";

  export let closeModal: () => void;
  export let entityType: EntityType;
  export let currentAssetName: string;

  const queryClient = useQueryClient();

  let error: string;

  $: allNamesQuery = useAllNames();

  const renameAsset = useRuntimeServiceRenameFileAndReconcile();

  const { form, errors, handleSubmit } = createForm({
    initialValues: {
      newName: currentAssetName,
    },
    validationSchema: yup.object({
      newName: yup
        .string()
        .matches(
          /^[a-zA-Z_][a-zA-Z0-9_]*$/,
          "Name must start with a letter or underscore and contain only letters, numbers, and underscores"
        )
        .required("Enter a name!")
        .notOneOf([currentAssetName], `That's the current name!`),
    }),
    onSubmit: async (values) => {
      if (
        isDuplicateName(values.newName, currentAssetName, $allNamesQuery.data)
      ) {
        error = `Name ${values.newName} is already in use`;
        return;
      }
      try {
        await renameFileArtifact(
          queryClient,
          currentAssetName,
          values.newName,
          entityType,
          $renameAsset
        );
        goto(getRouteFromName(values.newName, entityType), {
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
