import { type V1Expression } from "@rilldata/web-common/runtime-client";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

type YAMLConfigProviderEvents = {
  /** Fired after `update` has applied the config from a spec. */
  update: void;
};

/**
 * A provider for YAML only configuration. These are only mutable during yaml editing.
 */
export class YAMLConfigProvider {
  public defaultFilters = $state<Record<string, V1Expression | undefined>>({});
  // Lists, not sets: the filter bar shows the chips in the order the YAML lists them.
  // Filters toggled in the editor are added at the end. The `spec` lists hold what the file holds.
  public pinnedFilters = $state<string[]>([]);
  public specPinnedFilters = $state<string[]>([]);
  public requiredFilters = $state<string[]>([]);
  public specRequiredFilters = $state<string[]>([]);
  // Dimensions whose filter holds one value at a time. See `DimensionFilterManager.isSingleSelect`.
  public singleSelectFilters = $state<Record<string, boolean>>({});
  public editable = $state<boolean>(false);

  public cleanup: (() => void) | undefined = undefined;

  private events = new EventEmitter<YAMLConfigProviderEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  // The config last applied by `update`.
  private appliedConfig: string | undefined = undefined;

  /**
   * Applies the filter config of a dashboard's spec.
   * This resets the pin and required toggles made in the editor to what the spec holds,
   * so a config that is already applied is skipped:
   * the spec is fetched again on every edit to the dashboard's file, and unsaved toggles must survive unrelated edits.
   */
  public update(
    defaultFilters: Record<string, V1Expression | undefined>,
    pinnedFilters: string[],
    requiredFilters: string[],
    singleSelectFilters: string[] = [],
  ) {
    const config = JSON.stringify([
      defaultFilters,
      pinnedFilters,
      requiredFilters,
      singleSelectFilters,
    ]);
    if (config === this.appliedConfig) return;
    this.appliedConfig = config;

    this.defaultFilters = defaultFilters;

    this.pinnedFilters = [...new Set(pinnedFilters)];
    this.specPinnedFilters = [...this.pinnedFilters];

    this.requiredFilters = [...new Set(requiredFilters)];
    this.specRequiredFilters = [...this.requiredFilters];

    this.singleSelectFilters = Object.fromEntries(
      singleSelectFilters.map((filter) => [filter, true]),
    );

    this.events.emit("update");
  }

  /**
   * Whether `name` is listed under `filters.single_select`.
   * Compares to `true`, since a plain lookup also finds inherited properties like `constructor`.
   */
  public isSingleSelectFilter(name: string) {
    return this.singleSelectFilters[name] === true;
  }

  public setEditable(newEditable: boolean) {
    this.editable = newEditable;
  }

  public isPinnedFilter(name: string) {
    return this.pinnedFilters.includes(name);
  }

  public isRequiredFilter(name: string) {
    return this.requiredFilters.includes(name);
  }

  public isPinnedOrRequiredFilter(name: string) {
    return this.isPinnedFilter(name) || this.isRequiredFilter(name);
  }

  public togglePinnedFilter(filter: string) {
    this.pinnedFilters = toggle(this.pinnedFilters, filter);
  }

  public toggleRequiredFilter(filter: string) {
    this.requiredFilters = toggle(this.requiredFilters, filter);
  }
}

// Adds the name at the end of the list, or removes it.
function toggle(names: string[], name: string) {
  return names.includes(name)
    ? names.filter((n) => n !== name)
    : [...names, name];
}
