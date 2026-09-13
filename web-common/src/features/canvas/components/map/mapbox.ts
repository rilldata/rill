// The default token is a public (`pk.*`) Mapbox token restricted to the styles
// used below. It is meant to ship in the client bundle and can be rotated from
// the Mapbox console without a code change. Deployments that want to use their
// own Mapbox account can override it at build time by setting
// `RILL_UI_PUBLIC_MAPBOX_ACCESS_TOKEN` (see `envPrefix` in the Vite configs).
const DEFAULT_MAPBOX_ACCESS_TOKEN =
  "pk.eyJ1IjoicmlsbGRhdGEiLCJhIjoiY21nemp4Mnl3MDViaGQzc2J0MzB1NjdvMiJ9.4Q8jXek0-EF4RLA_TF4-oA";

export const MAPBOX_ACCESS_TOKEN: string =
  (import.meta.env.RILL_UI_PUBLIC_MAPBOX_ACCESS_TOKEN as string | undefined) ??
  DEFAULT_MAPBOX_ACCESS_TOKEN;

export const MAPBOX_STYLE_LIGHT = "mapbox://styles/mapbox/light-v11";
export const MAPBOX_STYLE_DARK = "mapbox://styles/mapbox/dark-v11";
