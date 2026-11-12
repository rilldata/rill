import { goto } from "$app/navigation";
import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import { derived, get, readable, writable, type Readable } from "svelte/store";
import {
  collectControls,
  CONTROL_URL_PREFIX,
  controlsForPath,
  lockedControlValue,
  parseControlValue,
  serializeControlValue,
  type AppliedControl,
  type LockedControls,
  type ResolvedControl,
} from "./controls";

/**
 * Holds the viewer controls of a canvas and their values. The URL is the source of truth:
 * each control that isn't on its default has a `c.<scope>.<name>` search param.
 */
export class ControlsManager {
  /** Every control on the canvas, with its current values. */
  readonly all: Readable<AppliedControl[]>;
  /** Controls whose value differs from their default. */
  readonly changed: Readable<AppliedControl[]>;
  /** URL keys of controls whose URL value wasn't allowed, so their default applies. */
  readonly invalid: Readable<string[]>;

  /** URL key of the control the viewer is pointing at, so the widgets it changes can show it. */
  readonly hovered = writable<string | null>(null);

  /** The URL params that hold control values. Empty while controls are disabled. */
  private readonly searchParams: Readable<URLSearchParams>;

  constructor(
    private readonly spec: Readable<V1CanvasSpec | undefined>,
    searchParams: Readable<URLSearchParams>,
    /** Whether viewers can change controls. When false, every control stays on its default. */
    readonly enabled: Readable<boolean> = readable(true),
    /**
     * Called with the new search string before each navigation. The canvas restores its
     * last-visited state when the URL has no params, which would bring back the values of
     * controls that were just reset, so it must learn the new state first.
     */
    private readonly onNavigate: (search: string) => void = () => {},
    /** Values the embedding page fixed. They win over the URL, and viewers can't change them. */
    private readonly locked: LockedControls = {},
  ) {
    const noParams = new URLSearchParams();
    this.searchParams = derived(
      [searchParams, enabled],
      ([$params, $enabled]) => ($enabled ? $params : noParams),
    );
    searchParams = this.searchParams;
    this.all = derived([spec, searchParams], ([$spec, $params]) =>
      collectControls($spec).map((c) => apply(c, $params, locked)),
    );
    this.changed = derived([spec, searchParams], ([$spec, $params]) =>
      collectControls($spec)
        .filter((c) => {
          if (lockedControlValue(locked, c.urlKey) !== null) return false;
          const raw = $params.get(c.urlKey);
          if (raw === null) return false;
          const { values, valid } = parseControlValue(c.control, raw);
          return (
            valid && serializeControlValue(c.control, values) !== undefined
          );
        })
        .map((c) => apply(c, $params, locked)),
    );
    this.invalid = derived([spec, searchParams], ([$spec, $params]) =>
      collectControls($spec)
        .filter(
          (c) =>
            lockedControlValue(locked, c.urlKey) === null &&
            !parseControlValue(c.control, $params.get(c.urlKey)).valid,
        )
        .map((c) => c.urlKey),
    );
  }

  /**
   * The controls in effect for the item at a YAML path, nearest level first.
   * The path is read on every evaluation, since an item's path changes when rows move.
   */
  forPath(getPath: () => (string | number)[]): Readable<AppliedControl[]> {
    return derived([this.spec, this.searchParams], ([$spec, $params]) =>
      controlsForPath($spec, getPath()).map((c) =>
        apply(c, $params, this.locked),
      ),
    );
  }

  /** Sets a control's values. Null resets it to its default. */
  set(control: ResolvedControl, values: string[] | null) {
    if (typeof window === "undefined") return;
    if (lockedControlValue(this.locked, control.urlKey) !== null) return;
    const params = new URLSearchParams(window.location.search);
    const raw = serializeControlValue(control.control, values);
    // An empty string is a value: the viewer cleared a control whose default isn't empty.
    if (raw === undefined) params.delete(control.urlKey);
    else params.set(control.urlKey, raw);
    this.navigate(params);
  }

  /** Resets every control to its default. Other URL state is left as it is. */
  resetAll() {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    for (const { urlKey } of get(this.all)) params.delete(urlKey);
    this.navigate(params);
  }

  private navigate(params: URLSearchParams) {
    const search = params.toString();
    this.onNavigate(search);
    goto(`?${search}`).catch(console.error);
  }
}

function apply(
  c: ResolvedControl,
  params: URLSearchParams,
  locked: LockedControls,
): AppliedControl {
  const lockedValue = lockedControlValue(locked, c.urlKey);
  return {
    ...c,
    values: parseControlValue(c.control, lockedValue ?? params.get(c.urlKey))
      .values,
    locked: lockedValue !== null,
  };
}

export { CONTROL_URL_PREFIX };
