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
  import { get } from "svelte/store";
  import type { YAMLPath } from "../../layout-util";
  import { slotFields, specUsesField } from "../apply-fields";
  import {
    controlToYAML,
    defaultControlName,
    filterBarTimeRanges,
    findLevelConflict,
    findRepeatedOption,
    listAffectedItems,
    listFieldTargets,
    TIME_GRAIN_OPTIONS,
    type ControlDraft,
    type ControlLevelRef,
    type EditableControlKind,
  } from "../control-edit";
  import { controlOptionLabel } from "../control-labels";
  import { ControlKind, offeredTimeGrains } from "../controls";
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

  const NAME_REGEX = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

  // The form starts from the props as they are when the dialog opens; later changes to them
  // (e.g. the document reconciling) must not reset what the builder typed.
  const seed = untrack(() => ({ initial, levels, levelPath }));

  // Time controls offer what the canvas's filter bar offers: its time ranges, and the grains its
  // metrics views offer.
  const timePresets = untrack(() => {
    const { yamlConfigProvider, metricsViewsProvider } =
      canvas.dashboardProvider;
    return {
      ranges: filterBarTimeRanges(
        yamlConfigProvider.timeRanges,
        metricsViewsProvider.largestMinTimeGrain,
        metricsViewsProvider.maxQueryTimeRange,
      ),
      grains: offeredTimeGrains(
        TIME_GRAIN_OPTIONS,
        metricsViewsProvider.availableTimeGrains,
      ),
    };
  });

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
          .filter((o) => !timePresets.ranges.includes(o))
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

  // The widget the dialog's item level holds, if it offers one: viewers can pick the fields of
  // its properties that hold a list of them, such as a KPI grid's measures (a slot control).
  let itemLevel = $derived(levels.find((l) => l.kind === "item"));
  let components = $derived(canvas.componentsStore);
  let component = $derived.by(() => {
    if (!itemLevel) return undefined;
    const key = itemLevel.path.join("/");
    return [...$components.values()].find(
      (c) => c.pathInYAML.slice(0, -1).join("/") === key,
    );
  });
  let slotTargets = $derived(
    component
      ? listFieldTargets(component.inputParams(component.type).options)
      : [],
  );

  let isSwap = $derived(
    kind === ControlKind.Measure || kind === ControlKind.Dimension,
  );
  let isSlot = $derived(kind === ControlKind.Slot);
  let isFilter = $derived(kind === ControlKind.Filter);
  // The filter bar's `single_select` keeps a dimension's filter to one value, and its controls too.
  let singleSelect = $derived(
    isFilter &&
      !!field &&
      canvas.dashboardProvider.yamlConfigProvider.isSingleSelectFilter(field),
  );
  let needsField = $derived(isSwap || isFilter || isSlot);
  // A slot control belongs to its widget.
  let level = $derived(
    isSlot && itemLevel ? itemLevel : (levels[Number(levelIndex)] ?? levels[0]),
  );
  let slotTarget = $derived(
    isSlot ? slotTargets.find((t) => t.target === field) : undefined,
  );

  // Measures or dimensions of the metrics views in play, by name.
  let fields = $derived.by(() => {
    const res = new Map<string, string>();
    const views = metricsViewName
      ? [$metricsViews[metricsViewName]]
      : Object.values($metricsViews);
    const withMeasures =
      kind === ControlKind.Measure || !!slotTarget?.allowed.includes("measure");
    const withDimensions =
      kind === ControlKind.Dimension ||
      isFilter ||
      !!slotTarget?.allowed.includes("dimension");
    for (const view of views) {
      const spec = view?.state?.validSpec;
      const list = [
        ...(withDimensions ? (spec?.dimensions ?? []) : []),
        ...(withMeasures ? (spec?.measures ?? []) : []),
      ];
      for (const f of list) {
        if (f.name && !res.has(f.name))
          res.set(f.name, f.displayName || f.name);
      }
    }
    // Keep a field the control already names, even if no metrics view has it now.
    const named = isSlot ? options : [field, ...(isSwap ? options : [])];
    for (const f of named) {
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
    if (isSlot)
      return fieldOptions
        .map((o) => o.value)
        .filter((v) => options.includes(v));
    if (isSwap) {
      const picked = options.filter((o) => o !== field && fields.has(o));
      return field ? [field, ...picked] : picked;
    }
    if (kind === ControlKind.TimeRange) {
      return [
        ...timePresets.ranges.filter((o) => options.includes(o)),
        ...splitList(customRangesText),
      ];
    }
    return timePresets.grains.filter((o) => options.includes(o));
  });

  let checklist = $derived(
    isSwap || isSlot
      ? fieldOptions
      : (kind === ControlKind.TimeRange
          ? timePresets.ranges
          : timePresets.grains
        ).map((value) => ({
          value,
          label: controlOptionLabel({ kind }, value, (name) => name),
        })),
  );

  // The name the control will have, which items' `ignore_controls` refer to.
  let controlName = $derived(
    name.trim() ||
      defaultControlName(
        controlToYAML({
          kind,
          field,
          options: [],
          default: [],
          multiple: false,
          display: "",
          label: "",
          name: "",
        }),
      ),
  );
  // Whether a widget's component spec is one the control changes: a swap needs the field in it,
  // a filter the dimension in its metrics view, and a time control a time dimension there.
  function changesSpec(spec: Record<string, unknown>): boolean {
    if (isSwap) return specUsesField(spec, field);
    const view =
      typeof spec.metrics_view === "string"
        ? $metricsViews[spec.metrics_view]?.state?.validSpec
        : undefined;
    if (isFilter) return !!view?.dimensions?.some((d) => d.name === field);
    return !!view?.timeDimension;
  }
  // The widgets the control would change at the chosen level, listed so the designer sees
  // which ones follow it. A slot control only ever changes its own widget.
  let affectedItems = $derived(
    !isSlot && (!needsField || field)
      ? listAffectedItems(
          $parsedContent,
          level.path,
          kind,
          field,
          controlName,
          changesSpec,
        )
      : null,
  );
  let affected = $derived(affectedItems?.length ?? null);
  const MAX_LISTED_ITEMS = 4;
  let allAffectedNames = $derived(
    affectedItems?.map((i) => i.title).join(", ") ?? "",
  );
  let affectedNames = $derived.by(() => {
    if (!affectedItems?.length) return "";
    const titles = affectedItems.map((i) => `“${i.title}”`);
    const listed = titles.slice(0, MAX_LISTED_ITEMS).join(", ");
    return titles.length > MAX_LISTED_ITEMS
      ? `${listed} ${m.dashboard_others({ count: titles.length - MAX_LISTED_ITEMS })}`
      : listed;
  });

  // Another control on the level that changes the same thing. Only the first of two applies,
  // so the runtime rejects the second.
  let originalName = $derived(
    seed.initial ? defaultControlName(controlToYAML(seed.initial)) : "",
  );
  let levelConflict = $derived(
    !needsField || field
      ? findLevelConflict(
          $parsedContent,
          level.path,
          kind,
          field,
          !isNew && level.path.join("/") === seed.levelPath.join("/")
            ? originalName
            : "",
        )
      : "",
  );
  // A swap to a field a widget already shows would show it twice. The runtime rejects that control
  // on the widget itself; under a control of an enclosing level, the widget keeps its own field.
  let repeated = $derived(
    isSwap && field
      ? findRepeatedOption(
          $parsedContent,
          level.path,
          kind,
          field,
          resolvedOptions,
          controlName,
        )
      : null,
  );

  let nameError = $derived(
    name && !NAME_REGEX.test(name)
      ? m.canvas_control_dialog_invalid_name()
      : "",
  );
  let optionsError = $derived.by(() => {
    if ((isSwap || isSlot) && resolvedOptions.length < 2) {
      return m.canvas_control_dialog_need_two_options();
    }
    if (!isSwap && !isFilter && resolvedOptions.length < 1) {
      return m.canvas_control_dialog_need_option();
    }
    return "";
  });
  let onItem = $derived(level.kind === "item");
  let valid = $derived(
    (!needsField || !!field) &&
      !nameError &&
      !optionsError &&
      !(onItem && repeated) &&
      !levelConflict &&
      // On a widget, a swap of a field it doesn't show would change nothing.
      !(onItem && affected === 0),
  );

  function setKind(next: string) {
    kind = next as EditableControlKind;
    // Fields and options of one kind mean nothing to another.
    field = "";
    options = [];
    defaultValue = "";
    if (kind === ControlKind.Slot && slotTargets[0]) {
      setSlotTarget(slotTargets[0].target);
    }
  }

  // Viewers start from the fields the widget shows, so those are offered by default.
  function setSlotTarget(target: string) {
    field = target;
    options = component ? slotFields(get(component.specStore), target) : [];
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
        multiple: (isFilter && multiple && !singleSelect) || isSlot,
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
          ...(slotTargets.length || isSlot ? [ControlKind.Slot] : []),
          ControlKind.Filter,
          ControlKind.TimeRange,
          ControlKind.TimeGrain,
        ].map((value) => ({ value, label: controlKindLabel(value) }))}
        onChange={setKind}
      />

      {#if isSlot}
        <Select
          id="canvas-control-slot"
          label={m.canvas_control_dialog_slot_target()}
          size="md"
          sameWidth
          full
          value={field}
          options={slotTargets.map((t) => ({
            value: t.target,
            label: t.label,
          }))}
          onChange={setSlotTarget}
        />
      {:else if needsField}
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
        <div class="flex flex-col gap-y-1">
          <div class="flex items-center gap-x-2">
            <Switch
              id="canvas-control-multiple"
              small
              checked={multiple && !singleSelect}
              disabled={singleSelect}
              onCheckedChange={(checked) => (multiple = checked)}
            />
            <label for="canvas-control-multiple" class="text-sm">
              {m.canvas_control_dialog_multiple()}
            </label>
          </div>
          {#if singleSelect}
            <span class="text-xs text-fg-secondary">
              {m.canvas_control_dialog_multiple_single_select()}
            </span>
          {/if}
        </div>
      {:else}
        <div class="flex flex-col gap-y-1">
          <InputLabel
            id="canvas-control-options"
            capitalize={false}
            label={isSlot
              ? m.canvas_control_dialog_options_slot()
              : kind === ControlKind.Measure
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
            <!-- Saving waits for this, so it reads as an error. -->
            <span class="text-xs text-destructive" role="alert">
              {optionsError}
            </span>
          {/if}
          {#if isSlot}
            <span class="text-xs text-fg-secondary">
              {m.canvas_control_dialog_slot_hint()}
            </span>
          {/if}
          {#if repeated && onItem}
            <span class="text-xs text-destructive" role="alert">
              {m.canvas_control_dialog_repeated({
                field: fields.get(repeated.option) ?? repeated.option,
                item: repeated.item,
              })}
            </span>
          {:else if repeated}
            <span class="text-xs text-fg-secondary">
              {m.canvas_control_dialog_repeated_kept({
                field: fields.get(repeated.option) ?? repeated.option,
                item: repeated.item,
              })}
            </span>
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

        {#if !isSwap && !isSlot}
          <Select
            id="canvas-control-default"
            label={m.canvas_control_dialog_default()}
            size="md"
            sameWidth
            full
            value={resolvedOptions.includes(defaultValue) ? defaultValue : ""}
            placeholder={m.canvas_inherit_from_canvas()}
            clearable
            options={resolvedOptions.map((value) => ({
              value,
              label: controlOptionLabel({ kind }, value, (name) => name),
            }))}
            onChange={(value) => (defaultValue = value)}
          />
        {/if}
      {/if}

      {#if levels.length > 1 && !isSlot}
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

      {#if levelConflict}
        <p class="text-xs text-destructive" role="alert">
          {m.canvas_control_dialog_level_conflict({
            level: controlLevelLabel(level),
            name: levelConflict,
          })}
        </p>
      {:else if affected !== null}
        <p
          class="text-xs"
          class:text-fg-secondary={affected > 0}
          title={allAffectedNames || undefined}
        >
          {affected > 0
            ? m.canvas_control_dialog_affects({
                count: affected,
                names: affectedNames,
              })
            : isSwap
              ? m.canvas_control_dialog_affects_none()
              : isFilter
                ? m.canvas_control_dialog_affects_none_filter()
                : m.canvas_control_dialog_affects_none_time()}
        </p>
      {/if}

      <div class="grid gap-x-3" class:grid-cols-2={!isSlot}>
        <!-- The Select next to it puts its label further from its field. -->
        <Input
          id="canvas-control-label"
          capitalizeLabel={false}
          size="md"
          labelGap={2}
          label={m.canvas_control_dialog_label()}
          optional
          bind:value={label}
        />
        {#if !isSlot}
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
        {/if}
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
