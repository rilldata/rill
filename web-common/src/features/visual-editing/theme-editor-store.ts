import type {
  V1ThemeSpec,
  V1ThemeColors,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { Theme } from "../themes/theme";
import { themePreviewOverride } from "../themes/selectors";

interface ThemeEditorState {
  mode: "presets" | "custom";
  editing: boolean;
  selectedThemeName: string | undefined;
  baseSpec: V1ThemeSpec | undefined;
  // Edits keyed by variable name, per light/dark mode
  lightEdits: Record<string, string>;
  darkEdits: Record<string, string>;
}

const initialState: ThemeEditorState = {
  mode: "presets",
  editing: false,
  selectedThemeName: undefined,
  baseSpec: undefined,
  lightEdits: {},
  darkEdits: {},
};

function createThemeEditorStore() {
  const store = writable<ThemeEditorState>({ ...initialState });

  function getBaseValues(
    spec: V1ThemeSpec | undefined,
    isDark: boolean,
  ): Record<string, string> {
    if (!spec) return {};
    const modeColors: V1ThemeColors | undefined = isDark
      ? spec.dark
      : spec.light;
    if (!modeColors) return {};

    const result: Record<string, string> = {};
    if (modeColors.primary) result.primary = modeColors.primary;
    if (modeColors.secondary) result.secondary = modeColors.secondary;
    if (modeColors.variables) {
      Object.assign(result, modeColors.variables);
    }
    return result;
  }

  function buildSpecFromState(state: ThemeEditorState): V1ThemeSpec {
    const base = state.baseSpec ?? {};
    const lightBase = getBaseValues(base, false);
    const darkBase = getBaseValues(base, true);

    const mergedLight = { ...lightBase, ...state.lightEdits };
    const mergedDark = { ...darkBase, ...state.darkEdits };

    return {
      light: extractThemeColors(mergedLight),
      dark: extractThemeColors(mergedDark),
    };
  }

  function extractThemeColors(
    flat: Record<string, string>,
  ): V1ThemeColors | undefined {
    const { primary, secondary, ...rest } = flat;
    const variables = Object.keys(rest).length > 0 ? rest : undefined;
    if (!primary && !secondary && !variables) return undefined;
    return { primary, secondary, variables };
  }

  function updatePreview(state: ThemeEditorState) {
    if (!state.editing) {
      themePreviewOverride.set(undefined);
      return;
    }
    const spec = buildSpecFromState(state);
    themePreviewOverride.set(new Theme(spec));
  }

  return {
    subscribe: store.subscribe,

    startEditing(spec: V1ThemeSpec, themeName: string | undefined) {
      const state: ThemeEditorState = {
        mode: "presets",
        editing: true,
        selectedThemeName: themeName,
        baseSpec: spec,
        lightEdits: { ...getBaseValues(spec, false) },
        darkEdits: { ...getBaseValues(spec, true) },
      };
      store.set(state);
      updatePreview(state);
    },

    startCustom(spec: V1ThemeSpec) {
      const state: ThemeEditorState = {
        mode: "custom",
        editing: true,
        selectedThemeName: undefined,
        baseSpec: spec,
        lightEdits: { ...getBaseValues(spec, false) },
        darkEdits: { ...getBaseValues(spec, true) },
      };
      store.set(state);
      updatePreview(state);
    },

    updateProperty(key: string, value: string, isDarkMode: boolean) {
      store.update((s) => {
        if (isDarkMode) {
          s.darkEdits = { ...s.darkEdits, [key]: value };
        } else {
          s.lightEdits = { ...s.lightEdits, [key]: value };
        }
        updatePreview(s);
        return s;
      });
    },

    /** Returns the current merged spec for YAML serialization on save */
    buildSpec(): V1ThemeSpec {
      return buildSpecFromState(get(store));
    },

    /** Returns the flat map of values for the given mode, merging base + edits */
    getValues(isDarkMode: boolean): Record<string, string> {
      const state = get(store);
      const base = getBaseValues(state.baseSpec, isDarkMode);
      const edits = isDarkMode ? state.darkEdits : state.lightEdits;
      return { ...base, ...edits };
    },

    markSaved() {
      store.update((s) => {
        const newSpec = buildSpecFromState(s);
        s.baseSpec = newSpec;
        updatePreview(s);
        return s;
      });
    },

    exitEditing() {
      store.set({ ...initialState });
      themePreviewOverride.set(undefined);
    },
  };
}

export const themeEditorStore = createThemeEditorStore();
