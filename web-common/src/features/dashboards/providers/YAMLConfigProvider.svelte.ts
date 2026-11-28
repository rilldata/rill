import {
  type V1ExploreTimeRange,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import { DEFAULT_TIMEZONE } from "@rilldata/web-common/lib/time/config.ts";
import { arrayOrderedEquals } from "@rilldata/web-common/lib/arrayUtils.ts";

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

  public restrictedDimensions = $state<string[] | undefined>(undefined);
  public primaryTimeDimension = $state<string | undefined>(undefined);
  public restrictedMeasures = $state<string[] | undefined>(undefined);

  public defaultTimeRange = $state<string | undefined>(undefined);
  public timeRanges = $state<V1ExploreTimeRange[]>([]);
  public allowCustomTimeRange = $state(false);
  public defaultTimeZone = $state<string>(DEFAULT_TIMEZONE);
  public timeZones = $state<string[]>([]);
  public lockTimeZone = $state(false);

  public editable = $state<boolean>(false);

  public cleanup: (() => void) | undefined = undefined;

  private events = new EventEmitter<YAMLConfigProviderEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  // The part of the config last applied by `update` that the editor does not change.
  private appliedSpecConfigJson: string | undefined = undefined;

  /**
   * Applies the config of a dashboard's spec.
   * The spec is fetched again on every edit to the dashboard's file,
   * and the pin and required toggles made in the editor must survive edits that do not touch them.
   * So each of those lists is reset to what the spec holds only when the spec's list changes,
   * and the rest of the config is applied only when it changes.
   *
   * Note: any changes to editors that writes temporary fields before yaml write
   *       need to make sure they are checked separately than specConfig.
   */
  public update(config: {
    defaultFilters?: YAMLConfigProvider["defaultFilters"];
    pinnedFilters?: string[];
    requiredFilters?: string[];
    singleSelectFilters?: string[];

    restrictedDimensions?: YAMLConfigProvider["restrictedDimensions"];
    primaryTimeDimension?: YAMLConfigProvider["primaryTimeDimension"];
    restrictedMeasures?: YAMLConfigProvider["restrictedMeasures"];

    defaultTimeRange?: YAMLConfigProvider["defaultTimeRange"];
    timeRanges?: YAMLConfigProvider["timeRanges"];
    allowCustomTimeRange?: YAMLConfigProvider["allowCustomTimeRange"];
    defaultTimeZone?: YAMLConfigProvider["defaultTimeZone"];
    timeZones?: YAMLConfigProvider["timeZones"];
    lockTimeZone?: YAMLConfigProvider["lockTimeZone"];
  }) {
    const { pinnedFilters, requiredFilters, ...specConfig } = config;
    let changed = false;

    const specPinnedFilters = [...new Set(pinnedFilters)];
    if (!arrayOrderedEquals(specPinnedFilters, this.specPinnedFilters)) {
      this.specPinnedFilters = specPinnedFilters;
      this.pinnedFilters = [...specPinnedFilters];
      changed = true;
    }

    const specRequiredFilters = [...new Set(requiredFilters)];
    if (!arrayOrderedEquals(specRequiredFilters, this.specRequiredFilters)) {
      this.specRequiredFilters = specRequiredFilters;
      this.requiredFilters = [...specRequiredFilters];
      changed = true;
    }

    const specConfigJson = JSON.stringify(specConfig);
    if (specConfigJson !== this.appliedSpecConfigJson) {
      this.appliedSpecConfigJson = specConfigJson;
      this.applySpecConfig(specConfig);
      changed = true;
    }

    if (changed) this.events.emit("update");
  }

  private applySpecConfig({
    defaultFilters,
    singleSelectFilters,

    restrictedDimensions,
    primaryTimeDimension,
    restrictedMeasures,

    defaultTimeRange,
    timeRanges,
    allowCustomTimeRange,
    defaultTimeZone,
    timeZones,
    lockTimeZone,
  }: Omit<
    Parameters<YAMLConfigProvider["update"]>[0],
    "pinnedFilters" | "requiredFilters"
  >) {
    this.defaultFilters = defaultFilters ?? {};
    this.singleSelectFilters = Object.fromEntries(
      singleSelectFilters?.map((filter) => [filter, true]) ?? [],
    );

    this.restrictedDimensions = restrictedDimensions;
    this.primaryTimeDimension = primaryTimeDimension;
    this.restrictedMeasures = restrictedMeasures;

    this.defaultTimeRange = defaultTimeRange;
    this.timeRanges = timeRanges ?? [];
    this.allowCustomTimeRange = allowCustomTimeRange ?? false;
    this.defaultTimeZone = defaultTimeZone ?? DEFAULT_TIMEZONE;
    this.timeZones = timeZones ?? [];
    this.lockTimeZone = lockTimeZone ?? false;
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
