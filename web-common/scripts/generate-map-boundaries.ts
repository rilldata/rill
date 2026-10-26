/**
 * Generates the built-in region boundary sets used by the canvas map's region layer.
 *
 * Run manually when changing a boundary set or bumping the source data:
 *
 *   npm run generate:map-boundaries -w web-common
 *
 * Sources are public domain Natural Earth files, pinned to a release tag.
 * Each set is normalized to the properties the region matcher understands
 * (see web-common/src/features/canvas/components/map/regions/registry.ts),
 * then simplified and written as TopoJSON with mapshaper, which shares arcs between neighboring regions
 * and is several times smaller than the equivalent GeoJSON.
 * mapshaper is pinned below and fetched with npx when the script runs,
 * rather than added to web-common's dependencies, which would add ~170 packages (including a native addon) to every install.
 *
 * TODO: Move this script into its own package with mapshaper as a dependency,
 * so its version is managed in a package.json without weighing down web-common.
 *
 * TODO: The output is bundled into web-common for now. Move it to the CDN and load it from there
 * once the boundary sets grow (e.g. admin-1 for every country).
 */

import { execFileSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const NATURAL_EARTH_TAG = "v5.1.2";
const NATURAL_EARTH_BASE = `https://raw.githubusercontent.com/nvkelso/natural-earth-vector/${NATURAL_EARTH_TAG}/geojson`;
const MAPSHAPER = "mapshaper@0.7.98";
const OUTPUT_DIR = join(
  import.meta.dirname,
  "../src/features/canvas/components/map/regions/data",
);

type Properties = Record<string, string | number | null | undefined>;
type Feature = { type: "Feature"; properties: Properties; geometry: unknown };
type FeatureCollection = { type: "FeatureCollection"; features: Feature[] };

/**
 * Common names that Natural Earth does not carry for a country, keyed by ISO 3166-1 alpha-2 code.
 * Matching ignores case, accents and punctuation, so "U.S.A." and "usa" don't need separate entries.
 */
const COUNTRY_ALIASES: Record<string, string[]> = {
  BN: ["Brunei Darussalam"],
  BO: ["Bolivia, Plurinational State of"],
  CD: [
    "DRC",
    "DR Congo",
    "Congo, Democratic Republic of the",
    "Congo-Kinshasa",
  ],
  CG: ["Congo", "Republic of the Congo", "Congo-Brazzaville"],
  CI: ["Cote d'Ivoire"],
  CV: ["Cabo Verde"],
  CZ: ["Czech Republic"],
  FM: ["Micronesia, Federated States of"],
  GB: ["UK", "Great Britain", "Britain"],
  IR: ["Iran, Islamic Republic of"],
  KP: ["Korea, Democratic People's Republic of", "DPRK"],
  KR: ["Korea", "Korea, Republic of"],
  LA: ["Lao PDR", "Lao People's Democratic Republic"],
  MD: ["Moldova, Republic of"],
  MK: ["Macedonia", "North Macedonia"],
  MM: ["Burma"],
  PS: ["Palestine, State of"],
  RU: ["Russian Federation"],
  SY: ["Syrian Arab Republic"],
  TL: ["East Timor"],
  TR: ["Turkiye", "Türkiye"],
  TW: ["Taiwan, Province of China"],
  TZ: ["Tanzania, United Republic of"],
  US: ["USA", "US", "United States", "America"],
  VE: ["Venezuela, Bolivarian Republic of"],
  VN: ["Viet Nam"],
};

const US_STATE_ALIASES: Record<string, string[]> = {
  "US-DC": ["Washington DC", "Washington, D.C.", "D.C."],
};

async function main() {
  const countries = await fetchCollection("ne_50m_admin_0_countries");
  writeBoundarySet("countries", {
    type: "FeatureCollection",
    features: countries.features.map((f) => {
      const p = f.properties;
      const iso2 = isoCode(p.ISO_A2_EH);
      return {
        ...f,
        properties: {
          id: p.ADM0_A3,
          name: p.NAME,
          iso_a2: iso2,
          iso_a3: isoCode(p.ISO_A3_EH),
          iso_n3: isoCode(p.ISO_N3_EH),
          aliases: joinAliases(p.NAME as string, [
            p.NAME_LONG,
            p.ADMIN,
            p.FORMAL_EN,
            p.NAME_EN,
            p.NAME_SORT,
            ...String(p.NAME_ALT ?? "").split("|"),
            ...(iso2 ? (COUNTRY_ALIASES[iso2] ?? []) : []),
          ]),
        },
      };
    }),
  });

  const states = await fetchCollection("ne_50m_admin_1_states_provinces");
  writeBoundarySet("us_states", {
    type: "FeatureCollection",
    features: states.features
      .filter((f) => f.properties.adm0_a3 === "USA")
      .map((f) => {
        const p = f.properties;
        const iso = p.iso_3166_2 as string;
        return {
          ...f,
          properties: {
            id: iso,
            name: p.name,
            postal: p.postal,
            iso_3166_2: iso,
            // Natural Earth prefixes the two digit FIPS state code with the country, e.g. "US06".
            fips: String(p.fips).replace(/^US/, ""),
            aliases: joinAliases(p.name as string, [
              ...String(p.name_alt ?? "").split("|"),
              ...(US_STATE_ALIASES[iso] ?? []),
            ]),
          },
        };
      }),
  });
}

async function fetchCollection(name: string): Promise<FeatureCollection> {
  const res = await fetch(`${NATURAL_EARTH_BASE}/${name}.geojson`);
  if (!res.ok) throw new Error(`failed to download ${name}: ${res.status}`);
  return (await res.json()) as FeatureCollection;
}

function writeBoundarySet(name: string, collection: FeatureCollection) {
  const workDir = mkdtempSync(join(tmpdir(), "rill-map-boundaries-"));
  try {
    const input = join(workDir, `${name}.geojson`);
    writeFileSync(input, JSON.stringify(collection));
    execFileSync(
      "npx",
      [
        "-y",
        MAPSHAPER,
        "-i",
        input,
        "-simplify",
        "20%",
        "keep-shapes",
        "-rename-layers",
        name,
        "-o",
        join(OUTPUT_DIR, `${name}.topo.json`),
        "format=topojson",
        "id-field=id",
        "quantization=100000",
      ],
      {
        stdio: "inherit",
        // The repo's .npmrc sets prefer-offline, which resolves mapshaper's dependencies from cached registry metadata;
        // when that cache predates a dependency's release, npx fails with "No matching version found".
        env: { ...process.env, npm_config_prefer_offline: "false" },
      },
    );
  } finally {
    rmSync(workDir, { recursive: true, force: true });
  }
}

/** Natural Earth marks a missing code as "-99". */
function isoCode(value: Properties[string]): string | undefined {
  const code = String(value ?? "");
  return code && code !== "-99" ? code : undefined;
}

/** Aliases are stored as one "|" separated string, since TopoJSON tooling does not reliably keep array properties. */
function joinAliases(name: string, values: unknown[]): string {
  const aliases = new Set<string>();
  for (const v of values) {
    if (typeof v === "string" && v.trim() && v !== name) aliases.add(v.trim());
  }
  return [...aliases].join("|");
}

await main();
