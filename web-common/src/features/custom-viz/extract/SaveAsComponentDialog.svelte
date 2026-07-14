<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { CustomChartComponent } from "@rilldata/web-common/features/canvas/components/charts/custom-chart";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { extractCustomChart, type LiftedField } from "./extract-component";

  export let open = false;
  export let component: CustomChartComponent;

  const client = useRuntimeClient();

  let name = "";
  let saving = false;
  let liftField: Record<string, boolean> = {};

  $: componentsQuery = useFilteredResources(client, ResourceKind.Component);
  $: existingNames = ($componentsQuery?.data ?? [])
    .map((res) => res.meta?.name?.name)
    .filter((existing): existing is string => !!existing);

  // Seed a unique default name when the dialog opens.
  $: if (open && !name) {
    name = getName("custom_viz", existingNames);
  }

  $: fields = get(component.queryFieldsMeta);
  $: timeDimension = getTimeDimension();

  function getTimeDimension(): string | undefined {
    const store = component.parent.metricsView.getTimeDimensionForMetricView(
      component.metricsViewName,
    );
    return get(store);
  }

  $: nameError = validateName(name);

  function validateName(candidate: string): string | undefined {
    if (!candidate) return m.component_extract_name_required();
    if (!/^[a-zA-Z0-9_-]+$/.test(candidate)) {
      return m.component_extract_name_invalid();
    }
    if (existingNames.some((n) => n.toLowerCase() === candidate.toLowerCase())) {
      return m.component_extract_name_taken();
    }
    return undefined;
  }

  async function save() {
    if (nameError || saving) return;
    saving = true;
    try {
      const lifted: LiftedField[] = fields
        .filter((field) => liftField[field.name])
        .map((field) => ({
          ...field,
          isTimeDimension:
            field.type === "dimension" && field.name === timeDimension,
        }));

      await extractCustomChart(
        client,
        component,
        name,
        prettyDisplayName(name),
        lifted,
      );

      eventBus.emit("notification", {
        message: m.component_extract_success({ name }),
      });
      open = false;
      name = "";
      liftField = {};
    } catch (error) {
      eventBus.emit("notification", {
        message: extractErrorMessage(error),
        type: "error",
      });
    } finally {
      saving = false;
    }
  }

  function prettyDisplayName(value: string): string {
    return value
      .replace(/[_-]/g, " ")
      .replace(/(^|\s)\w/g, (c) => c.toUpperCase());
  }
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>
      {m.component_extract_title()}
    </AlertDialog.Title>

    <AlertDialog.Description>
      {m.component_extract_description()}
    </AlertDialog.Description>

    <Input
      label={m.component_extract_name_label()}
      id="component-name"
      bind:value={name}
      errors={nameError ? [nameError] : undefined}
    />

    {#if fields.length > 0}
      <div class="flex flex-col gap-y-1">
        <div class="text-xs font-medium text-fg-secondary">
          {m.component_extract_fields_label()}
        </div>
        <div class="text-[11px] text-fg-secondary pb-1">
          {m.component_extract_fields_note()}
        </div>
        {#each fields as field (field.name)}
          <Checkbox
            label={`${field.display_name ?? field.name} (${field.type})`}
            bind:checked={liftField[field.name]}
          />
        {/each}
      </div>
    {/if}

    <AlertDialog.Footer>
      <AlertDialog.Cancel>
        {#snippet child({ props })}
          <Button {...props} large type="secondary">{m.common_cancel()}</Button>
        {/snippet}
      </AlertDialog.Cancel>

      <AlertDialog.Action>
        {#snippet child({ props })}
          <Button
            {...props}
            disabled={!!nameError || saving}
            large
            type="primary"
            onClick={save}
          >
            {m.component_extract_cta()}
          </Button>
        {/snippet}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
