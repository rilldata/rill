<script lang="ts">
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import {
    GCS,
    GCSYupSchema,
    HTTP,
    HTTPYupSchema,
    S3,
    S3YupSchema,
  } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import Input from "../../Input.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";
  import Tab from "../../tab/Tab.svelte";
  import TabGroup from "../../tab/TabGroup.svelte";

  let selectedConnector = "S3";
  let connectorSpec = S3;
  let yupSchema = S3YupSchema;

  function extendYupSchemaWithSourceName(yupSchema: yup.AnyObjectSchema) {
    return yupSchema.concat(
      yup.object().shape({
        sourceName: yup.string().required(),
      })
    );
  }

  yupSchema = extendYupSchemaWithSourceName(yupSchema);

  function onConnectorChange(selectedConnector: string) {
    if (selectedConnector === "S3") {
      connectorSpec = S3;
      yupSchema = S3YupSchema;
    } else if (selectedConnector === "GCS") {
      connectorSpec = GCS;
      yupSchema = GCSYupSchema;
    } else if (selectedConnector === "HTTP") {
      connectorSpec = HTTP;
      yupSchema = HTTPYupSchema;
    }

    yupSchema = extendYupSchemaWithSourceName(yupSchema);
  }

  $: onConnectorChange(selectedConnector);

  function compileCreateSourceSql(values) {
    const compiledKeyValues = Object.entries(values)
      .map(([key, value]) => `${key}='${value}'`)
      .join(", ");

    return (
      `CREATE SOURCE ${values.sourceName} WITH (connector = '${connectorSpec.name}', ` +
      compiledKeyValues +
      `)`
    );
  }

  const { form, errors, handleSubmit } = createForm({
    // TODO: initialValues should come from SQL asset and be reactive to asset modifications
    initialValues: {},
    validationSchema: yupSchema,
    onSubmit: (values) => {
      const sql = compileCreateSourceSql(values);
      // TODO: dispatch sql to SQL editor
      // TODO: submit sql to backend
      alert(sql);
    },
  });
</script>

<div class="h-full flex flex-col">
  <TabGroup
    variant="secondary"
    on:select={(event) => {
      selectedConnector = event.detail;
    }}
  >
    <Tab value={"S3"}>S3</Tab>
    <Tab value={"GCS"}>GCS</Tab>
    <Tab value={"HTTP"}>https</Tab>
  </TabGroup>

  {#key selectedConnector}
    <div class="pt-8 flex-grow overflow-y-auto">
      <div>{@html connectorSpec.description}</div>
      <form
        on:submit={handleSubmit}
        id="remote-source-{selectedConnector}-form"
      >
        <div class="py-4">
          <Input
            label="Source name"
            bind:value={$form["sourceName"]}
            error={$errors["sourceName"]}
            placeholder="my_new_source"
          />
          {#each Object.entries(connectorSpec.fields) as [name, attributes]}
            {@const label =
              attributes.label + (attributes.required ? "" : " (optional)")}
            <div class="py-2">
              {#if attributes.type === "text"}
                <Input
                  id={name}
                  {label}
                  error={$errors[name]}
                  bind:value={$form[name]}
                  placeholder={attributes.placeholder}
                />
              {/if}
              {#if attributes.type === "checkbox"}
                <label for={name} class="flex items-center">
                  <input
                    id={name}
                    type="checkbox"
                    bind:checked={$form[name]}
                    class="h-5 w-5"
                  />
                  <span class="ml-2 text-sm">{label}</span>
                </label>
              {/if}
            </div>
          {/each}
        </div>
      </form>
    </div>
    <div class="">
      <DialogFooter>
        <Button
          type="primary"
          submitForm
          form="remote-source-{selectedConnector}-form"
        >
          Add source
        </Button>
      </DialogFooter>
    </div>
  {/key}
</div>
