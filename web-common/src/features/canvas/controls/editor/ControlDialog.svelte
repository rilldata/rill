<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { untrack } from "svelte";
  import type { YAMLPath } from "../../layout-util";
  import {
    countItemsUsingField,
    TIME_GRAIN_OPTIONS,
    type ControlDraft,
    type ControlLevelRef,
    type EditableControlKind,
  } from "../control-edit";
  import { controlOptionLabel } from "../control-labels";
  import { ControlKind } from "../controls";
  import { controlKindLabel, controlLevelLabel } from "./labels";

  // The form for one viewer control. It's open while mounted.
  let {
    canvas,
    levels,
    levelPath,
    initial = undefined,
    isNew,
    metricsViewName = undefined,
    onSave,
    onCancel,
  }: {
    canvas: CanvasEntity;
    // The levels the control can be declared on. With one level, the choice is hidden.
    levels: ControlLevelRef[];
    // The level the control is (or will be) declared on.
    levelPath: YAMLPath;
    initial?: ControlDraft;
    isNew: boolean;
    // Limits the fields on offer to one metrics view, e.g. a widget's.
    metricsViewName?: string;
    onSave: (draft: ControlDraft, levelPath: YAMLPath) => void;
    onCancel: () => void;
  } = $props();

  const TIME_RANGE_PRESETS = [
    "PT6H",
    "PT24H",
    "P7D",
    "P14D",
    "P4W",
    "P3M",
    "P6M",
    "P12M",
  ];
  const NAME_REGEX = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

  // The form starts from the props as they are when the dialog opens; later changes to them
  // (e.g. the document reconciling) must not reset what the builder typed.
  const seed = untrack(() => ({ initial, levels, levelPath }));

  let kind = $state<EditableControlKind>(
    seed.initial?.kind ?? ControlKind.Measure,
  );
  let field = $state(seed.initial?.field ?? "");
  let options = $state<string[]>(seed.initial?.options ?? []);
  let valuesText = $state(
    seed.initial?.kind === ControlKind.Filter
      ? seed.initial.options.join(", ")
      : "",
  );
  let customRangesText = $state(
    seed.initial?.kind === ControlKind.TimeRange
      ? seed.initial.options
          .filter((o) => !TIME_RANGE_PRESETS.includes(o))
          .join(", ")
      : "",
  );
  let defaultText = $state(
    seed.initial?.kind === ControlKind.Filter
      ? seed.initial.default.join(", ")
      : "",
  );
  let defaultValue = $state(
    seed.initial && seed.initial.kind !== ControlKind.Filter
      ? (seed.initial.default[0] ?? "")
      : "",
  );
  let multiple = $state(seed.initial?.multiple ?? false);
  let display = $state<string>(seed.initial?.display || "auto");
  let label = $state(seed.initial?.label ?? "");
  let name = $state(seed.initial?.name ?? "");
  let levelIndex = $state(
    String(
      Math.max(
        0,
        seed.levels.findIndex(
          (l) => l.path.join("/") === seed.levelPath.join("/"),
        ),
      ),
    ),
  );

  let metricsViews = $derived(canvas._metricsViews);
  let parsedContent = $derived(canvas.parsedContent);

  let isSwap = $derived(
    kind === ControlKind.Measure || kind === ControlKind.Dimension,
  );
  let isFilter = $derived(kind === ControlKind.Filter);
  let needsField = $derived(isSwap || isFilter);
  let level = $derived(levels[Number(levelIndex)] ?? levels[0]);

  // Measures or dimensions of the metrics views in play, by name.
  let fields = $derived.by(() => {
    const res = new Map<string, string>();
    const views = metricsViewName
      ? [$metricsViews[metricsViewName]]
      : Object.values($metricsViews);
    for (const view of views) {
      const spec = view?.state?.validSpec;
      const list =
        kind === ControlKind.Measure ? spec?.measures : spec?.dimensions;
      for (const f of list ?? []) {
        if (f.name && !res.has(f.name))
          res.set(f.name, f.displayName || f.name);
      }
    }
    // Keep a field the control already names, even if no metrics view has it now.
    for (const f of [field, ...(isSwap ? options : [])]) {
      if (f && !res.has(f)) res.set(f, f);
    }
    return res;
  });
  let fieldOptions = $derived(
    [...fields].map(([value, label]) => ({ value, label })),
  );

  function splitList(text: string): string[] {
    return [
      ...new Set(
        text
          .split(",")
          .map((v) => v.trim())
          .filter(Boolean),
      ),
    ];
  }

  // The options as they'd be written, per kind.
  let resolvedOptions = $derived.by(() => {
    if (isFilter) return splitList(valuesText);
    if (isSwap) {
      const picked = options.filter((o) => o !== field && fields.has(o));
      return field ? [field, ...picked] : picked;
    }
    if (kind === ControlKind.TimeRange) {
      return [
        ...TIME_RANGE_PRESETS.filter((o) => options.includes(o)),
        ...splitList(customRangesText),
      ];
    }
    return TIME_GRAIN_OPTIONS.filter((o) => options.includes(o));
  });

  let checklist = $derived(
    isSwap
      ? fieldOptions
      : (kind === ControlKind.TimeRange
          ? TIME_RANGE_PRESETS
          : TIME_GRAIN_OPTIONS
        ).map((value) => ({
          value,
          label: controlOptionLabel({ kind }, value, (name) => name),
        })),
  );

  let affected = $derived(
    isSwap && field
      ? countItemsUsingField($parsedContent, level.path, field)
      : null,
  );

  let nameError = $derived(
    name && !NAME_REGEX.test(name)
      ? m.canvas_control_dialog_invalid_name()
      : "",
  );
  let optionsError = $derived.by(() => {
    if (isSwap && resolvedOptions.length < 2) {
      return m.canvas_control_dialog_need_two_options();
    }
    if (!isSwap && !isFilter && resolvedOptions.length < 1) {
      return m.canvas_control_dialog_need_option();
    }
    return "";
  });
  let valid = $derived((!needsField || !!field) && !nameError && !optionsError);

  function setKind(next: string) {
    kind = next as EditableControlKind;
    // Fields and options of one kind mean nothing to another.
    field = "";
    options = [];
    defaultValue = "";
  }

  function toggleOption(value: string, checked: boolean) {
    options = checked
      ? [...options, value]
      : options.filter((o) => o !== value);
    if (!checked && defaultValue === value) defaultValue = "";
  }

  function save() {
    if (!valid) return;
    onSave(
      {
        kind,
        field: needsField ? field : "",
        options: resolvedOptions,
        default: isFilter
          ? splitList(defaultText)
          : defaultValue
            ? [defaultValue]
            : [],
        multiple: isFilter && multiple,
        display: display === "buttons" || display === "dropdown" ? display : "",
        label: label.trim(),
        name: name.trim(),
      },
      level.path,
    );
  }
</script>

<Dialog.Root
  open
  onOpenChange={(open) => {
    if (!open) onCancel();
  }}
>
  <Dialog.Content class="max-w-lg" aria-label={m.canvas_viewer_controls()}>
    <Dialog.Header>
      <Dialog.Title>
        {isNew
          ? m.canvas_control_dialog_add_title()
          : m.canvas_control_dialog_edit_title()}
      </Dialog.Title>
    </Dialog.Header>

    <div class="form">
      <Select
        id="canvas-control-kind"
        label={m.canvas_control_dialog_kind()}
        size="md"
        sameWidth
        full
        value={kind}
        options={[
          ControlKind.Measure,
          ControlKind.Dimension,
          ControlKind.Filter,
          ControlKind.TimeRange,
          ControlKind.TimeGrain,
        ].map((value) => ({ value, label: controlKindLabel(value) }))}
        onChange={setKind}
      />

      {#if needsField}
        <Select
          id="canvas-control-field"
          label={kind === ControlKind.Measure
            ? m.canvas_control_dialog_field_measure()
            : kind === ControlKind.Dimension
              ? m.canvas_control_dialog_field_dimension()
              : m.canvas_control_dialog_field_filter()}
          size="md"
          sameWidth
          full
          enableSearch
          bind:value={field}
          options={fieldOptions}
        />
      {/if}

      {#if isFilter}
        <Input
          id="canvas-control-values"
          capitalizeLabel={false}
          size="md"
          label={m.canvas_control_dialog_allowed_values()}
          hint={m.canvas_control_dialog_allowed_values_hint()}
          optional
          bind:value={valuesText}
        />
        <Input
          id="canvas-control-default-values"
          capitalizeLabel={false}
          size="md"
          label={m.canvas_control_dialog_default()}
          hint={m.canvas_control_dialog_default_filter_hint()}
          optional
          bind:value={defaultText}
        />
        <div class="flex items-center gap-x-2">
          <Switch id="canvas-control-multiple" small bind:checked={multiple} />
          <label for="canvas-control-multiple" class="text-sm">
            {m.canvas_control_dialog_multiple()}
          </label>
        </div>
      {:else}
        <div class="flex flex-col gap-y-1">
          <InputLabel
            id="canvas-control-options"
            capitalize={false}
            label={kind === ControlKind.Measure
              ? m.canvas_control_dialog_options_measure()
              : kind === ControlKind.Dimension
                ? m.canvas_control_dialog_options_dimension()
                : kind === ControlKind.TimeRange
                  ? m.canvas_control_dialog_options_time_range()
                  : m.canvas_control_dialog_options_time_grain()}
          />
          <div
            class="checklist"
            role="group"
            aria-label={m.canvas_control_dialog_options()}
          >
            {#each checklist as option (option.value)}
              {@const isDefaultField = isSwap && option.value === field}
              <Checkbox
                id="canvas-control-option-{option.value}"
                label={option.label}
                labelClass="text-xs"
                checked={isDefaultField || options.includes(option.value)}
                disabled={isDefaultField}
                onCheckedChange={(checked) =>
                  toggleOption(option.value, checked === true)}
              />
            {:else}
              <span class="text-xs text-fg-secondary">
                {m.canvas_control_dialog_no_fields()}
              </span>
            {/each}
          </div>
          {#if optionsError}
            <span class="text-xs text-fg-secondary">{optionsError}</span>
          {/if}
        </div>

        {#if kind === ControlKind.TimeRange}
          <Input
            id="canvas-control-custom-ranges"
            capitalizeLabel={false}
            size="md"
            label={m.canvas_control_dialog_custom_ranges()}
            hint={m.canvas_control_dialog_custom_ranges_hint()}
            optional
            bind:value={customRangesText}
          />
        {/if}

        {#if !isSwap}
          <Select
            id="canvas-control-default"
            label={m.canvas_control_dialog_default()}
            size="md"
            sameWidth
            full
            value={resolvedOptions.includes(defaultValue) ? defaultValue : ""}
            placeholder={m.canvas_control_same_as_canvas()}
            clearable
            options={resolvedOptions.map((value) => ({
              value,
              label: controlOptionLabel({ kind }, value, (name) => name),
            }))}
            onChange={(value) => (defaultValue = value)}
          />
        {/if}
      {/if}

      {#if levels.length > 1}
        <Select
          id="canvas-control-level"
          label={m.canvas_control_dialog_applies_to()}
          size="md"
          sameWidth
          full
          bind:value={levelIndex}
          options={levels.map((l, i) => ({
            value: String(i),
            label: controlLevelLabel(l),
          }))}
        />
      {/if}

      {#if affected !== null}
        <p class="text-xs" class:text-fg-secondary={affected > 0}>
          {affected > 0
            ? m.canvas_control_dialog_affects({ count: affected })
            : m.canvas_control_dialog_affects_none()}
        </p>
      {/if}

      <div class="grid grid-cols-2 gap-x-3">
        <Input
          id="canvas-control-label"
          capitalizeLabel={false}
          size="md"
          label={m.canvas_control_dialog_label()}
          optional
          bind:value={label}
        />
        <Select
          id="canvas-control-display"
          label={m.canvas_control_dialog_display()}
          size="md"
          sameWidth
          full
          bind:value={display}
          options={[
            { value: "auto", label: m.canvas_control_dialog_display_auto() },
            {
              value: "buttons",
              label: m.canvas_control_dialog_display_buttons(),
            },
            {
              value: "dropdown",
              label: m.canvas_control_dialog_display_dropdown(),
            },
          ]}
        />
      </div>

      <Input
        id="canvas-control-name"
        capitalizeLabel={false}
        size="md"
        fontFamily="var(--font-mono, monospace)"
        label={m.canvas_control_dialog_name()}
        hint={m.canvas_control_dialog_name_hint()}
        errors={nameError || undefined}
        alwaysShowError
        optional
        bind:value={name}
      />
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={onCancel}>{m.common_cancel()}</Button>
      <Button type="primary" disabled={!valid} onClick={save}>
        {isNew
          ? m.canvas_viewer_controls_add()
          : m.canvas_control_dialog_save()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .form {
    @apply flex flex-col gap-y-3 max-h-[65vh] overflow-y-auto px-0.5 py-1;
  }

  .checklist {
    @apply flex flex-col gap-y-1.5 max-h-40 overflow-y-auto;
    @apply rounded-sm border p-2;
  }
</style>
