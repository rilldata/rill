import type {
  V1ExploreTimeRange,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  DEFAULT_TIMEZONE,
  DEFAULT_TIMEZONES,
} from "@rilldata/web-common/lib/time/config.ts";
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
  public pinnedFilters = $state<Record<string, boolean>>({});
  public specPinnedFilters = $state<Record<string, boolean>>({});
  public requiredFilters = $state<Record<string, boolean>>({});
  public specRequiredFilters = $state<Record<string, boolean>>({});

  public restrictedDimensions = $state<string[] | undefined>(undefined);
  public primaryTimeDimension = $state<string | undefined>(undefined);
  public restrictedMeasures = $state<string[] | undefined>(undefined);

  public defaultTimeRange = $state<string | undefined>(undefined);
  public timeRanges = $state<V1ExploreTimeRange[]>([]);
  public allowCustomTimeRange = $state(false);
  public defaultTimeZone = $state<string>(DEFAULT_TIMEZONE);
  public timeZones = $state<string[]>(DEFAULT_TIMEZONES);

  public editable = $state<boolean>(false);

  public cleanup: (() => void) | undefined = undefined;

  private events = new EventEmitter<YAMLConfigProviderEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  public update({
    defaultFilters,
    pinnedFilters,
    requiredFilters,

    restrictedDimensions,
    primaryTimeDimension,
    restrictedMeasures,

    defaultTimeRange,
    timeRanges,
    allowCustomTimeRange,
    defaultTimeZone,
    timeZones,
  }: {
    defaultFilters?: YAMLConfigProvider["defaultFilters"];
    pinnedFilters?: string[];
    requiredFilters?: string[];

    restrictedDimensions?: YAMLConfigProvider["restrictedDimensions"];
    primaryTimeDimension?: YAMLConfigProvider["primaryTimeDimension"];
    restrictedMeasures?: YAMLConfigProvider["restrictedMeasures"];

    defaultTimeRange?: YAMLConfigProvider["defaultTimeRange"];
    timeRanges?: YAMLConfigProvider["timeRanges"];
    allowCustomTimeRange?: YAMLConfigProvider["allowCustomTimeRange"];
    defaultTimeZone?: YAMLConfigProvider["defaultTimeZone"];
    timeZones?: YAMLConfigProvider["timeZones"];
  }) {
    this.defaultFilters = defaultFilters ?? {};

    const pinnedFiltersRec = Object.fromEntries(
      pinnedFilters?.map((filter) => [filter, true]) ?? [],
    );
    this.pinnedFilters = { ...pinnedFiltersRec };
    this.specPinnedFilters = { ...pinnedFiltersRec };

    const requiredFiltersRec = Object.fromEntries(
      requiredFilters?.map((filter) => [filter, true]) ?? [],
    );
    this.requiredFilters = { ...requiredFiltersRec };
    this.specRequiredFilters = { ...requiredFiltersRec };

    this.restrictedDimensions = restrictedDimensions;
    this.primaryTimeDimension = primaryTimeDimension;
    this.restrictedMeasures = restrictedMeasures;

    this.defaultTimeRange = defaultTimeRange;
    this.timeRanges = timeRanges ?? [];
    this.allowCustomTimeRange = allowCustomTimeRange ?? false;
    this.defaultTimeZone = defaultTimeZone ?? DEFAULT_TIMEZONE;
    this.timeZones = timeZones ?? [];

    this.events.emit("update");
  }

  public setEditable(newEditable: boolean) {
    this.editable = newEditable;
  }

  public togglePinnedFilter(filter: string) {
    if (!this.pinnedFilters[filter]) {
      this.pinnedFilters[filter] = true;
    } else {
      delete this.pinnedFilters[filter];
    }
  }

  public toggleRequiredFilter(filter: string) {
    if (!this.requiredFilters[filter]) {
      this.requiredFilters[filter] = true;
    } else {
      delete this.requiredFilters[filter];
    }
  }

  public isPinnedOrRequiredFilter(name: string) {
    return Boolean(this.pinnedFilters[name] || this.requiredFilters[name]);
  }
}
