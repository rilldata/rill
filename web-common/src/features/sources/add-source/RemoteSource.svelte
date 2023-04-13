<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import DialogFooter from "@rilldata/web-common/components/modal/dialog/DialogFooter.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useSourceNames } from "@rilldata/web-common/features/sources/selectors";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    ConnectorProperty,
    ConnectorPropertyType,
    createRuntimeServiceDeleteFileAndReconcile,
    createRuntimeServicePutFileAndReconcile,
    V1Connector,
    V1ReconcileError,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import type * as yup from "yup";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { compileCreateSourceYAML, inferSourceName } from "../sourceUtils";
  import { createSource } from "./createSource";
  import { humanReadableErrorMessage } from "./errors";
  import {
    fromYupFriendlyKey,
    getYupSchema,
    toYupFriendlyKey,
  } from "./yupSchemas";

  export let connector: V1Connector;

  $: runtimeInstanceId = $runtime.instanceId;
  $: sourceNames = useSourceNames(runtimeInstanceId);

  const createSourceMutation = createRuntimeServicePutFileAndReconcile();
  let createSourceMutationError: {
    code: number;
    message: string;
  };
  $: createSourceMutationError = ($createSourceMutation?.error as any)?.response
    ?.data;
  const deleteSource = createRuntimeServiceDeleteFileAndReconcile();

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();

  let connectorProperties: ConnectorProperty[];
  let yupSchema: yup.AnyObjectSchema;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let touched: Writable<Record<any, boolean>>;
  let errors: Writable<Record<never, string>>;
  let handleChange: (event: Event) => any;
  let handleSubmit: (event: Event) => any;

  let waitingOnSourceImport = false;
  let error: V1ReconcileError;

  function onConnectorChange(connector: V1Connector) {
    yupSchema = getYupSchema(connector);

    ({ form, touched, errors, handleChange, handleSubmit } = createForm({
      // TODO: initialValues should come from SQL asset and be reactive to asset modifications
      initialValues: {
        sourceName: "", // avoids `values.sourceName` warning
      },
      validationSchema: yupSchema,
      onSubmit: async (values) => {
        overlay.set({ title: `Importing ${values.sourceName}` });
        const formValues = Object.fromEntries(
          Object.entries(values).map(([key, value]) => [
            fromYupFriendlyKey(key),
            value,
          ])
        );

        const yaml = compileCreateSourceYAML(formValues, connector.name);

        waitingOnSourceImport = true;
        try {
          const errors = await createSource(
            queryClient,
            runtimeInstanceId,
            values.sourceName,
            yaml,
            $createSourceMutation
          );
          error = errors[0];
          if (!error) {
            dispatch("close");
          } else {
            await deleteFileArtifact(
              queryClient,
              runtimeInstanceId,
              values.sourceName,
              EntityType.Table,
              $deleteSource,
              $appStore.activeEntity,
              $sourceNames.data,
              false
            );
          }
        } catch (err) {
          // no-op
        }
        waitingOnSourceImport = false;
        overlay.set(null);
      },
    }));

    // Place the "Source name" field directly under the "Path" field, which is the first property for each connector (s3, gcs, https).
    connectorProperties = [
      ...connector.properties.slice(0, 1),
      {
        key: "sourceName",
        displayName: "Source name",
        description: "The name of the source",
        placeholder: "my_new_source",
        type: ConnectorPropertyType.TYPE_STRING,
        nullable: false,
      },
      ...connector.properties.slice(1),
    ];
  }

  $: onConnectorChange(connector);

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($touched.sourceName) return;
      const sourceName = inferSourceName(connector, value);
      $form.sourceName = sourceName ? sourceName : $form.sourceName;
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  <form
    on:submit|preventDefault={handleSubmit}
    id="remote-source-{connector.name}-form"
    class="px-4 pb-2 flex-grow overflow-y-auto"
  >
    <div class="pt-4 pb-2">
      Need help? Refer to our
      <a
        href="https://docs.rilldata.com/using-rill/import-data"
        target="_blank"
        rel="noreferrer">docs</a
      > for more information.
    </div>
    {#if $createSourceMutation.isError || error}
      <SubmissionError
        message={humanReadableErrorMessage(
          connector.name,
          createSourceMutationError?.code ?? 3,
          createSourceMutationError?.message ?? error.message
        )}
      />
    {/if}

    {#each connectorProperties as property}
      {@const label =
        property.displayName + (property.nullable ? " (optional)" : "")}
      <div class="py-2">
        {#if property.type === ConnectorPropertyType.TYPE_STRING}
          <Input
            id={toYupFriendlyKey(property.key)}
            {label}
            placeholder={property.placeholder}
            hint={property.hint}
            error={$errors[toYupFriendlyKey(property.key)]}
            bind:value={$form[toYupFriendlyKey(property.key)]}
            on:input={onStringInputChange}
            on:change={handleChange}
          />
        {:else if property.type === ConnectorPropertyType.TYPE_BOOLEAN}
          <label for={property.key} class="flex items-center">
            <input
              id={property.key}
              type="checkbox"
              bind:checked={$form[property.key]}
              class="h-5 w-5"
            />
            <span class="ml-2 text-sm">{label}</span>
          </label>
        {:else if property.type === ConnectorPropertyType.TYPE_INFORMATIONAL}
          <InformationalField
            description={property.description}
            hint={property.hint}
            href={property.href}
          />
        {/if}
      </div>
    {/each}
  </form>
  <div class="bg-gray-100 border-t border-gray-300">
    <DialogFooter>
      <div class="flex items-center space-x-2">
        <Button
          type="primary"
          submitForm
          form="remote-source-{connector.name}-form"
          disabled={$createSourceMutation.isLoading || waitingOnSourceImport}
        >
          Add source
        </Button>
      </div>
    </DialogFooter>
  </div>
</div>
