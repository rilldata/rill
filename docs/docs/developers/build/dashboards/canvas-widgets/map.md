---
title: "Map Widget"
description: Plot a metrics view on a map by point or region in Rill Canvas dashboards
sidebar_label: "Map"
sidebar_position: 15
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

The map widget plots a metrics view on an interactive basemap. Use it to show how a measure is distributed geographically: stores, sensors, or customers as points, or states, countries, or sales territories as shaded regions. For more information, refer to our [Components reference doc](/reference/project-files/component#map).

A map has up to two layers:

- A **geo dimension** layer, which draws a location stored in every row as a point or polygon, colored by a measure. Points can also be sized by a second measure.
- A **regions** layer, which fills built-in country or US state boundaries by a measure, matched through the values of a regular dimension such as `country` or `state`. It needs no geographic data in your model. See [Built-in regions](#built-in-regions).

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/map-points.png"
  imageAlt="Point map of stores across the United States, colored by revenue and sized by orders"
  code={`- map:
      metrics_view: store_sales_metrics
      title: Revenue by store
      geo_dimension:
        field: location
        type: nominal
      color:
        measure: total_revenue
        colorRange:
          mode: scheme
          scheme: tealblues
      size_measure:
        field: total_orders
        type: quantitative
      tooltip_dimension:
        field: city
        type: nominal`}
  codeLanguage="yaml"
/>

## Prepare geographic data

The geo dimension must contain one of the following values in every row. Coordinates are always longitude first, then latitude, in degrees (WGS84).

| Format | Example | Draws |
|--------|---------|-------|
| GeoJSON geometry string | `{"type":"Point","coordinates":[-122.42,37.77]}` | `Point`, `MultiPoint`, `Polygon`, and `MultiPolygon` geometries |
| GeoJSON `Feature` string | `{"type":"Feature","geometry":{...},"properties":{}}` | The feature's geometry. Its properties are ignored. |
| DuckDB `POINT_2D` column | `ST_Point2D(-122.42, 37.77)` | A point |
| DuckDB `POLYGON_2D` column | A polygon made of `[longitude, latitude]` rings | A polygon |

Rows with an empty value, a string that isn't valid GeoJSON, or a line geometry aren't drawn.

:::tip Use the spatial extension
Rill loads the [DuckDB spatial extension](https://duckdb.org/docs/extensions/spatial/overview) by default, so functions such as `ST_Point`, `ST_AsGeoJSON`, and `ST_Read` work in any DuckDB model without extra setup.
:::

### Points from latitude and longitude columns

Convert latitude and longitude columns to a GeoJSON string in your model. Note that `ST_Point` takes longitude first. Cast the result to `VARCHAR` so it is stored as a string.

```sql
-- models/store_sales.sql
SELECT
  city,
  state,
  ST_AsGeoJSON(ST_Point(longitude, latitude))::VARCHAR AS location,
  orders,
  revenue
FROM stores
```

Alternatively, `ST_Point2D(longitude, latitude) AS location` stores a DuckDB `POINT_2D` value, which Rill detects as a geo dimension automatically.

### Regions from a GeoJSON file

For most regions, a regions layer is simpler: use [built-in regions](#built-in-regions) for countries and US states, or [custom regions](#custom-regions) for others, such as DMAs, postal codes, or sales territories. Alternatively, load the boundaries as polygons in a geo dimension. The following model reads a GeoJSON `FeatureCollection` file from your project's `data` folder and returns one row per feature, with the feature's geometry as a GeoJSON string:

```sql
-- models/state_population.sql
SELECT
  feature.properties.name AS state,
  to_json(feature.geometry)::VARCHAR AS boundary,
  feature.properties.density AS population_density
FROM (
  SELECT unnest(features) AS feature
  FROM read_json('data/us-states.geojson')
)
```

Relative paths in `read_json` resolve from your project root. For other spatial formats, such as shapefiles, `ST_Read` returns each shape in a `geom` column that you can convert with `ST_AsGeoJSON(geom)::VARCHAR`.

To color regions by data from another model, join your facts to the boundaries on a shared key, such as a state or country code.

### Mark the dimension as a geo dimension

In the metrics view, set `type: geo` on the dimension that holds the location. The visual editor only offers geo dimensions in the map's **Geo dimension** selector. Columns of type DuckDB `POINT_2D` or `POLYGON_2D`, or ClickHouse `Point` or `Polygon`, are detected as geo dimensions without setting the type.

```yaml
# metrics/store_sales_metrics.yaml
type: metrics_view
model: store_sales
dimensions:
  - name: location
    display_name: Location
    column: location
    type: geo
  - name: city
    display_name: City
    column: city
measures:
  - name: total_revenue
    display_name: Revenue
    expression: SUM(revenue)
    format_preset: currency_usd
  - name: total_orders
    display_name: Orders
    expression: SUM(orders)
```

See [Metrics view YAML](/reference/project-files/metrics-views) for all dimension properties.

## Add a map to a canvas

1. In your canvas dashboard, click **Add widget** and select **Map**.
2. Choose a **Metrics view**. The map selects the first geo dimension and the first measure of the metrics view.
3. Under **Geo dimension**, pick the dimension that holds the locations.
4. Under **Color**, pick the measure and a color scheme or gradient.
5. Optionally, add a **Size measure** and a **Tooltip dimension**.

To shade countries or US states instead, pick a dimension under **Regions**; see [Built-in regions](#built-in-regions).

The map draws as soon as it has a metrics view and either a geo dimension with a color measure or a complete regions layer. You can also define the widget directly in the canvas YAML, as in the examples on this page.

## Point maps

When the geo dimension holds points, each row is drawn as a circle. The color measure sets its color, and the optional `size_measure` scales its radius from 4 to 20 pixels between the smallest and largest value. Without a size measure, every point has the same size.

Use a point map for locations such as stores, venues, or devices. Nearby points can overlap at low zoom levels; zoom in to separate them.

## Region (choropleth) maps

When the geo dimension holds polygons, each region is filled with the color of its measure value and outlined in the theme's primary color. The map switches to region mode automatically based on the data, and hides the **Size measure** setting, which only applies to points.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/map-polygons.png"
  imageAlt="Map of US states shaded by population density"
  code={`- map:
      metrics_view: state_population_metrics
      title: Population density by state
      geo_dimension:
        field: boundary
        type: nominal
      color:
        measure: log_population_density
        colorRange:
          mode: scheme
          scheme: viridis
      tooltip_dimension:
        field: state
        type: nominal
      initial_view:
        longitude: -98.5
        latitude: 39.5
        zoom: 3`}
  codeLanguage="yaml"
/>

Colors are spread linearly between the lowest and highest value on the map. If a few regions have much larger values than the rest, most regions end up in the same color. In the example above, the measure is `LN(AVG(population_density))` for that reason; a log-scaled or ratio measure (such as a per-capita value) often spreads colors more evenly.

## Built-in regions

To shade countries or US states, add a `regions` layer. It matches the values of a regular dimension to built-in boundaries, so your model only needs a column with country or state names or codes:

```yaml
- map:
    metrics_view: sales_metrics
    title: Revenue by country
    regions:
      dimension: country
      boundaries: countries
      color:
        measure: total_revenue
        colorRange:
          mode: scheme
          scheme: blues
```

In the visual editor, pick the dimension under **Regions**, then choose the **Boundaries**. When you add a map to a metrics view without a geo dimension, the map starts with a regions layer for the first dimension whose name contains `country` or `state`.

| `boundaries` | Covers | `key` values |
|--------------|--------|--------------|
| `countries` | Every country | `iso_a2` (`US`), `iso_a3` (`USA`), `iso_n3` (`840`), `name` (`United States`) |
| `us_states` | The 50 US states and the District of Columbia | `postal` (`CA`), `iso_3166_2` (`US-CA`), `fips` (`06`), `name` (`California`) |

By default, values are matched against every key, so a column that mixes `US`, `USA`, and `United States` still works. Set `key` to match one key only. Matching ignores case, accents, and punctuation, and `name` also matches common alternative names, such as `UK` or `Ivory Coast`. The legend reports how many values didn't match a region; hover over the message to see them.

Regions without data are drawn in gray. Regions are drawn below the basemap's labels and below a geo dimension layer, so you can combine both, for example to show stores as points over states shaded by revenue. The `color`, `size_measure`, and `tooltip_dimension` properties only apply to the geo dimension layer; the regions layer has its own `color`.

Boundaries come from [Natural Earth](https://www.naturalearthdata.com/), which draws de facto borders for disputed territories. For other regions, such as DMAs, use [custom regions](#custom-regions).

## Custom regions

To shade regions that aren't built in, such as Nielsen DMAs, sales territories, or store trade areas, add a GeoJSON or TopoJSON boundary file to your project's `public` folder and point `boundaries` to it:

```yaml
- map:
    metrics_view: ad_metrics
    title: Impressions by DMA
    regions:
      dimension: dma_code
      boundaries:
        file: public/dma.topojson
        id: dma_code
        name: dma_name
      color:
        measure: impressions
```

| Property | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Path of the boundary file from the project root. It must be directly inside one of the project's [`public_paths`](/reference/project-files/rill-yaml#public_paths), which default to `public`; files in subfolders aren't served. |
| `id` | Yes | Feature property that identifies each region. The values of `dimension` are matched against it, ignoring case, punctuation, and leading zeros. Features without this property use their top-level GeoJSON `id` instead. |
| `name` | No | Feature property shown as the region name in tooltips. Defaults to the `id` property. |

The file must be a GeoJSON `FeatureCollection` or a TopoJSON file with a single object. Only `Polygon` and `MultiPolygon` features with an id are drawn; if several features share an `id`, the first one is used. Your metrics view only needs a dimension whose values match the `id` property; no joins or geo dimensions are required. `key` doesn't apply to custom regions.

In the visual editor, choose **Custom file** under **Boundaries**, then enter the file and properties.

Every viewer's browser downloads the whole file, so keep it small. Simplifying the shapes and converting them to TopoJSON, which shares borders between neighboring regions, typically makes the file several times smaller:

```sh
npx mapshaper dma.geojson -simplify 10% keep-shapes -o format=topojson public/dma.topojson
```

The file is served only to viewers of your project, who receive it in full, including any viewers of [embedded dashboards](/developers/embed). Shapes aren't subject to [security policies](/developers/build/metrics-view/security): regions a viewer can't see data for still draw, in gray. Boundary sets such as DMAs are often licensed; you are responsible for ensuring your license permits displaying the shapes to your dashboard viewers, including embed users.

## Properties

| Property | Required | Description |
|----------|----------|-------------|
| `metrics_view` | Yes | Metrics view to query. |
| `geo_dimension.field` | One of `geo_dimension` or `regions` | Dimension that holds each row's location. See [Prepare geographic data](#prepare-geographic-data). |
| `color.measure` | With `geo_dimension` | Measure that sets the color of each point or polygon. |
| `color.colorRange` | No | Color scale. Defaults to the `tealblues` scheme. See [Colors](#colors). |
| `size_measure.field` | No | Measure that scales the radius of each point. Points only. |
| `tooltip_dimension.field` | No | Dimension shown as the heading of the hover tooltip, such as a city or state name. |
| `regions` | One of `geo_dimension` or `regions` | Regions layer with `dimension`, `boundaries`, optional `key`, and `color` (`measure` and optional `colorRange`). See [Built-in regions](#built-in-regions) and [Custom regions](#custom-regions). |
| `initial_view` | No | `longitude`, `latitude`, and optional `zoom` the map opens with. See [Set the initial view](#set-the-initial-view). |
| `title`, `description` | No | Title and description shown in the widget header. |
| `time_filters`, `dimension_filters` | No | Filters that apply to this widget only, as on other [canvas widgets](/developers/build/dashboards/canvas#filters). |

The `type` values under `geo_dimension`, `size_measure`, and `tooltip_dimension` (`nominal` or `quantitative`) are written by the visual editor. You can include them as shown in the examples.

### Colors

`colorRange` takes one of two forms:

```yaml
# A named color scheme
color:
  measure: total_revenue
  colorRange:
    mode: scheme
    scheme: viridis

# A gradient between two colors
color:
  measure: total_revenue
  colorRange:
    mode: gradient
    start: "#dbeafe"
    end: primary
```

Available schemes are `tealblues`, `viridis`, `magma`, `inferno`, `plasma`, `cividis`, `blues`, `teals`, `greens`, `greys`, `oranges`, `purples`, `reds`, `turbo`, and `spectral`, plus `sequential` and `diverging`, which use the palettes of your [dashboard theme](/developers/build/dashboards/customization). Gradient `start` and `end` accept a hex color, or `primary` and `secondary` for the theme colors.

### Set the initial view

By default, the map zooms to fit all of its data the first time it loads. To open the map on a fixed area instead, set `initial_view`:

```yaml
initial_view:
  longitude: -98.5
  latitude: 39.5
  zoom: 3
```

`zoom` ranges from `0`, which shows the whole world, to about `22`, which shows individual buildings. When `initial_view` is set, the map never zooms to fit the data.

## Interacting with a map

- **Pan and zoom:** Drag to pan, scroll or pinch to zoom, and use the zoom and compass buttons in the top-right corner. A scale bar is shown in the bottom-left corner.
- **Tooltips:** Hover over a point or polygon to see its tooltip dimension value, the color measure, and the size measure. Hover over a region to see its name and value.
- **Legend:** The top-left corner shows the color scale of each layer.
- **Filters:** The map follows the canvas time range and filters. Changing a filter updates the data without moving the map, so the area you zoomed into stays in view.
- **Dark mode:** The basemap switches between a light and a dark style with the dashboard's theme mode.

## Limitations

- Coordinates must be longitude and latitude in degrees. Projected coordinates, such as meters in a local coordinate system, aren't supported. Reproject them in your model first, for example with DuckDB's `ST_Transform`.
- Line geometries (`LineString`, `MultiLineString`) aren't drawn.
- Each layer loads up to 5,000 rows. The geo dimension layer loads one row for each combination of geo dimension and tooltip dimension values. Choose a tooltip dimension that has one value per location, and aggregate very large point sets, for example to a grid or to regions, before mapping them.
- The basemap is loaded from [Mapbox](https://www.mapbox.com/) in the viewer's browser, so viewers need access to `api.mapbox.com`.

:::note Self-hosted builds
Rill includes a Mapbox access token for the basemap. If you build the Rill frontend yourself and want to use your own Mapbox account, set `RILL_UI_PUBLIC_MAPBOX_ACCESS_TOKEN` to a public (`pk.`) Mapbox token when you build it. The value is embedded in the frontend bundle at build time, so setting it on a running server has no effect.
:::
