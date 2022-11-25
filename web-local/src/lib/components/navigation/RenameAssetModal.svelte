<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceRenameFileAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { getLabel } from "@rilldata/web-local/lib/components/entity-mappers/mappers";
  import { renameEntity } from "@rilldata/web-local/lib/svelte-query/actions";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { runtimeStore } from "../../application-state-stores/application-store";
  import Input from "../forms/Input.svelte";
  import SubmissionError from "../forms/SubmissionError.svelte";
  import { Dialog } from "../modal/index";

  export let closeModal: () => void;
  export let entityType: EntityType;
  export let currentAssetName: string;

  let error: string;

  $: runtimeInstanceId = $runtimeStore.instanceId;
  $: getCatalog = useRuntimeServiceGetCatalogEntry(
    runtimeInstanceId,
    currentAssetName
  );

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
      try {
        await renameEntity(
          runtimeInstanceId,
          currentAssetName,
          values.newName,
          entityType,
          $renameAsset
        );
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
