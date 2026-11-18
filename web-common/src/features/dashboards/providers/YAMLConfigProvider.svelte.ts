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
  public pinnedFilters = $state<Record<string, boolean>>({});
  public specPinnedFilters = $state<Record<string, boolean>>({});
  public requiredFilters = $state<Record<string, boolean>>({});
  public specRequiredFilters = $state<Record<string, boolean>>({});
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
  ) {
    const config = JSON.stringify([
      defaultFilters,
      pinnedFilters,
      requiredFilters,
    ]);
    if (config === this.appliedConfig) return;
    this.appliedConfig = config;

    this.defaultFilters = defaultFilters;

    const pinnedFiltersRec = Object.fromEntries(
      pinnedFilters.map((filter) => [filter, true]),
    );
    this.pinnedFilters = { ...pinnedFiltersRec };
    this.specPinnedFilters = { ...pinnedFiltersRec };

    const requiredFiltersRec = Object.fromEntries(
      requiredFilters.map((filter) => [filter, true]),
    );
    this.requiredFilters = { ...requiredFiltersRec };
    this.specRequiredFilters = { ...requiredFiltersRec };

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
