export interface ThemePropertyDef {
  key: string;
  label: string;
}

export interface ThemeSection {
  id: string;
  label: string;
  properties: ThemePropertyDef[];
  defaultOpen?: boolean;
}

export const THEME_SECTIONS: ThemeSection[] = [
  {
    id: "core",
    label: "Core",
    defaultOpen: true,
    properties: [
      { key: "primary", label: "Primary" },
      { key: "secondary", label: "Secondary" },
    ],
  },
  {
    id: "surfaces",
    label: "Surfaces",
    properties: [
      { key: "surface-subtle", label: "Subtle" },
      { key: "surface-background", label: "Background" },
      { key: "surface-card", label: "Card" },
    ],
  },
  {
    id: "foreground",
    label: "Text / Foreground",
    properties: [{ key: "fg-primary", label: "Primary" }],
  },
  {
    id: "kpi",
    label: "KPI",
    properties: [
      { key: "kpi-positive", label: "Positive" },
      { key: "kpi-negative", label: "Negative" },
    ],
  },
  {
    id: "qualitative",
    label: "Qualitative Palette",
    properties: Array.from({ length: 24 }, (_, i) => ({
      key: `color-qualitative-${i + 1}`,
      label: `${i + 1}`,
    })),
  },
  {
    id: "sequential",
    label: "Sequential Palette",
    properties: Array.from({ length: 9 }, (_, i) => ({
      key: `color-sequential-${i + 1}`,
      label: `${i + 1}`,
    })),
  },
  {
    id: "diverging",
    label: "Diverging Palette",
    properties: Array.from({ length: 11 }, (_, i) => ({
      key: `color-diverging-${i + 1}`,
      label: `${i + 1}`,
    })),
  },
];
