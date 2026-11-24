/**
 * Generates the built-in region boundary sets used by the canvas map's region layer:
 * `countries`, and `states` for each country in STATE_COUNTRIES.
 *
 * The output is not checked in. It runs as part of web-local's and web-admin's `dev` and `build` scripts,
 * and can be run manually with:
 *
 *   npm run generate:map-boundaries -w web-common [-- --force]
 *
 * Sources are public domain Natural Earth files, pinned to a release tag and verified against their SHA-256 checksums.
 * Each set is normalized to the properties the region matcher understands
 * (see web-common/src/features/canvas/components/map/regions/registry.ts),
 * then simplified and written as TopoJSON with mapshaper, which shares arcs between neighboring regions
 * and is several times smaller than the equivalent GeoJSON.
 *
 * The output goes to web-common/static/map-boundaries, which both apps serve at /map-boundaries.
 * Each file name includes a hash of its content, so it can be cached indefinitely;
 * manifest.json maps sets such as `countries` and `states/DE` to file names.
 * The manifest also records a stamp of the inputs (this script and the mapshaper version);
 * when the stamp matches and every listed file exists, generation is skipped unless `--force` is passed.
 */

import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { createRequire } from "node:module";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import mapshaper from "mapshaper";

const NATURAL_EARTH_TAG = "v5.1.2";
const NATURAL_EARTH_BASE = `https://raw.githubusercontent.com/nvkelso/natural-earth-vector/${NATURAL_EARTH_TAG}/geojson`;
/** SHA-256 checksums of the source files at NATURAL_EARTH_TAG; update them when bumping the tag. */
const NATURAL_EARTH_CHECKSUMS: Record<string, string> = {
  ne_50m_admin_0_countries:
    "3e458fc036ad0a66411f2c1e6cac49c5d7bfb81cb1123bc513b22511a2b7fdeb",
  ne_10m_admin_1_states_provinces:
    "22d0e3ad85eb3e27f17cabf8ba2d50e554fbc27a87796ff891d958185da62fb5",
};
const DOWNLOAD_TIMEOUT_MS = 5 * 60 * 1000;
const OUTPUT_DIR = join(import.meta.dirname, "../static/map-boundaries");
const MANIFEST_FILE = join(OUTPUT_DIR, "manifest.json");

type Properties = Record<string, string | number | null | undefined>;
type Feature = { type: "Feature"; properties: Properties; geometry: unknown };
type FeatureCollection = { type: "FeatureCollection"; features: Feature[] };

/** Written to manifest.json; the frontend reads `files`. */
type Manifest = { stamp: string; files: Record<string, string> };

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

/**
 * Countries with a `states` boundary set, and the number of regions each is expected to have.
 * Natural Earth's admin-1 level is uneven between countries,
 * so each country is reviewed before it is added here:
 * its regions should be the country's first-level subdivisions in ISO 3166-2 (see regionOverride for the exceptions),
 * and the count catches upstream changes when the Natural Earth tag is bumped.
 * Left out because Natural Earth predates a reform of their subdivisions: Indonesia (2022), Norway (2024) and Vietnam (2025).
 */
const STATE_COUNTRIES: Record<string, number> = {
  // Americas
  US: 52,
  CA: 13,
  MX: 32,
  BR: 27,
  AR: 24,
  CL: 16,
  CO: 33,
  PE: 25,
  // Europe
  GB: 4,
  IE: 26,
  FR: 18,
  DE: 16,
  ES: 19,
  IT: 20,
  PT: 20,
  NL: 16,
  BE: 11,
  CH: 26,
  AT: 9,
  SE: 21,
  DK: 5,
  FI: 19,
  PL: 16,
  CZ: 14,
  GR: 14,
  RO: 42,
  HU: 43,
  TR: 81,
  UA: 27,
  RU: 83,
  // Asia and Oceania
  IN: 36,
  CN: 31,
  JP: 47,
  KR: 17,
  TH: 77,
  MY: 16,
  PH: 83,
  SG: 5,
  AU: 8,
  NZ: 17,
  // Middle East and Africa
  IL: 6,
  SA: 13,
  AE: 7,
  EG: 27,
  NG: 37,
  ZA: 9,
};

const ISO_3166_2 = /^[A-Z]{2}-[A-Z0-9]{1,3}$/;

/** Spain's autonomous communities and cities, keyed by Natural Earth's `region` property of their provinces. */
const SPAIN_COMMUNITIES: Record<string, { code: string; name: string }> = {
  Andalucía: { code: "ES-AN", name: "Andalucía" },
  Aragón: { code: "ES-AR", name: "Aragón" },
  Asturias: { code: "ES-AS", name: "Asturias" },
  "Canary Is.": { code: "ES-CN", name: "Canarias" },
  Cantabria: { code: "ES-CB", name: "Cantabria" },
  "Castilla y León": { code: "ES-CL", name: "Castilla y León" },
  "Castilla-La Mancha": { code: "ES-CM", name: "Castilla-La Mancha" },
  Cataluña: { code: "ES-CT", name: "Cataluña" },
  Ceuta: { code: "ES-CE", name: "Ceuta" },
  Extremadura: { code: "ES-EX", name: "Extremadura" },
  Galicia: { code: "ES-GA", name: "Galicia" },
  "Islas Baleares": { code: "ES-IB", name: "Illes Balears" },
  "La Rioja": { code: "ES-RI", name: "La Rioja" },
  Madrid: { code: "ES-MD", name: "Madrid" },
  Melilla: { code: "ES-ML", name: "Melilla" },
  Murcia: { code: "ES-MC", name: "Murcia" },
  "Foral de Navarra": { code: "ES-NC", name: "Navarra" },
  "País Vasco": { code: "ES-PV", name: "País Vasco" },
  Valenciana: { code: "ES-VC", name: "Comunitat Valenciana" },
};

/** The countries of the United Kingdom, keyed by Natural Earth's `geonunit` property of their local authorities. */
const UK_COUNTRIES: Record<string, { code: string; name: string }> = {
  England: { code: "GB-ENG", name: "England" },
  Scotland: { code: "GB-SCT", name: "Scotland" },
  Wales: { code: "GB-WLS", name: "Wales" },
  "Northern Ireland": { code: "GB-NIR", name: "Northern Ireland" },
};

/**
 * Names of regions that Natural Earth splits into several features,
 * where the first feature that is not a city does not carry the region's name.
 */
const MERGED_REGION_NAMES: Record<string, string> = {
  "IE-TA": "Tipperary",
  "PH-MNL": "Metro Manila",
};

/** Common names that Natural Earth does not carry for a region, keyed by ISO 3166-2 code. */
const REGION_ALIASES: Record<string, string[]> = {
  "US-DC": ["Washington DC", "Washington, D.C.", "D.C."],
  "GB-ENG": ["ENG"],
  "GB-SCT": ["SCT"],
  "GB-WLS": ["WLS", "Cymru"],
  "GB-NIR": ["NIR"],
  "FR-COR": ["Corsica"],
  "FR-GUF": ["French Guiana", "Guyane"],
  "ES-AN": ["Andalusia"],
  "ES-CN": ["Canary Islands", "Islas Canarias"],
  "ES-CT": ["Catalonia", "Catalunya"],
  "ES-IB": ["Balearic Islands", "Islas Baleares"],
  "ES-NC": ["Navarre", "Comunidad Foral de Navarra"],
  "ES-PV": ["Basque Country", "Euskadi"],
  "ES-VC": ["Valencian Community", "Comunidad Valenciana", "Valencia"],
  "IT-23": ["Aosta Valley", "Vallée d'Aoste"],
  "IT-32": ["Trentino-South Tyrol", "Trentino-Südtirol"],
  "IT-75": ["Puglia"],
  "IT-82": ["Sicilia"],
  "FI-01": ["Ahvenanmaa", "Aland Islands"],
  "CO-DC": ["Bogota D.C.", "Distrito Capital"],
};

async function main() {
  const stamp = inputStamp();
  if (!process.argv.includes("--force") && isUpToDate(stamp)) {
    console.log("Map boundaries are up to date; skipping generation.");
    return;
  }

  const files: Record<string, string> = {};

  const countries = await fetchCollection("ne_50m_admin_0_countries");
  files.countries = await toTopoJSON(
    "countries",
    {
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
    },
    ["-simplify 20% keep-shapes"],
  );

  const admin1 = await fetchCollection("ne_10m_admin_1_states_provinces");
  for (const [country, collection] of buildStates(admin1)) {
    files[`states/${country}`] = await toTopoJSON("states", collection, [
      // Merge the features of each region, then keep 10% of the vertices.
      // Unlike a fixed interval or a resolution, a percentage gives countries of every size a similar level of detail,
      // including those with distant overseas regions, such as France.
      "-dissolve id copy-fields=name,code,postal,iso_3166_2,fips,aliases",
      "-simplify 10% keep-shapes",
    ]);
  }

  writeOutput(stamp, files);
}

/**
 * Groups Natural Earth's admin-1 features into the regions of each country in STATE_COUNTRIES.
 * Features are grouped by ISO 3166-2 code, so a region that Natural Earth splits into several features,
 * such as a county and its city, is drawn as one;
 * each feature keeps its own geometry, and the features of a region are merged by mapshaper.
 * Countries follow the code's prefix, so regions are assigned to countries as in ISO 3166-2.
 * Features without a code are left out; they are small islands and disputed areas that data does not refer to.
 */
function buildStates(
  admin1: FeatureCollection,
): Map<string, FeatureCollection> {
  type Region = {
    code: string;
    name: string;
    namedAfterCity: boolean;
    aliases: unknown[];
    features: Feature[];
  };
  const regions = new Map<string, Region>();
  for (const f of admin1.features) {
    const p = f.properties;
    const override = regionOverride(p);
    const code = override?.code ?? String(p.iso_3166_2 ?? "");
    if (!ISO_3166_2.test(code) || !(code.slice(0, 2) in STATE_COUNTRIES)) {
      continue;
    }

    let region = regions.get(code);
    if (!region) {
      region = {
        code,
        name: "",
        namedAfterCity: false,
        aliases: [],
        features: [],
      };
      regions.set(code, region);
    }
    region.features.push(f);
    if (override) {
      region.name = override.name;
      continue;
    }
    // A region split into several features is named after its first feature that is not a city, such as a county.
    const isCity = /city/i.test(String(p.type_en ?? ""));
    if (!region.name || (region.namedAfterCity && !isCity)) {
      region.name = String(p.name);
      region.namedAfterCity = isCity;
    }
    region.aliases.push(
      p.name,
      p.name_en,
      ...String(p.name_alt ?? "").split("|"),
    );
  }

  const states = new Map<string, FeatureCollection>();
  for (const region of regions.values()) {
    const country = region.code.slice(0, 2);
    const name = MERGED_REGION_NAMES[region.code] ?? region.name;
    const properties: Properties = {
      id: region.code,
      name,
      code: region.code.slice(3),
      // Kept for the `us_states` boundary set, which matches the same codes by `postal`.
      postal: country === "US" ? region.code.slice(3) : undefined,
      iso_3166_2: region.code,
      // Natural Earth prefixes the FIPS state code with the country, and adds a county code for Puerto Rico: "US06", "72033".
      fips:
        country === "US"
          ? String(region.features[0].properties.fips)
              .replace(/^US/, "")
              .slice(0, 2)
          : undefined,
      aliases: joinAliases(name, [
        ...region.aliases,
        ...(REGION_ALIASES[region.code] ?? []),
      ]),
    };

    let collection = states.get(country);
    if (!collection) {
      collection = { type: "FeatureCollection", features: [] };
      states.set(country, collection);
    }
    for (const f of region.features) {
      collection.features.push({ ...f, properties });
    }
  }

  for (const [country, expected] of Object.entries(STATE_COUNTRIES)) {
    const codes = new Set(
      states.get(country)?.features.map((f) => f.properties.id),
    );
    if (codes.size !== expected) {
      throw new Error(
        `expected ${expected} regions for ${country}, found ${codes.size}: ${[...codes].sort().join(", ")}`,
      );
    }
  }
  return states;
}

/**
 * Code and name of the region that a Natural Earth admin-1 feature belongs to,
 * where its admin-1 level is below the first-level subdivisions in ISO 3166-2, or its code is wrong.
 * Returns undefined to use the feature's own `iso_3166_2` code.
 */
function regionOverride(
  p: Properties,
): { code: string; name: string } | undefined {
  switch (p.adm0_a3) {
    // Departments, merged into regions.
    case "FRA":
    // Provinces, merged into regions.
    // eslint-disable-next-line no-fallthrough
    case "ITA":
      return { code: String(p.region_cod).trim(), name: String(p.region) };
    // Provinces, merged into autonomous communities.
    // Natural Earth's codes for them are not ISO 3166-2 codes, and give Melilla the code of Ceuta, so they are looked up by name.
    case "ESP":
      return lookup(SPAIN_COMMUNITIES, p.region);
    // Local authorities, merged into the countries of the United Kingdom.
    case "GBR":
      return lookup(UK_COUNTRIES, p.geonunit);
    // Natural Earth treats Åland as a country, while ISO 3166-2 lists it as a region of Finland.
    case "ALD":
      return { code: "FI-01", name: "Åland" };
  }
  // Bogotá is a capital district with its own code, but Natural Earth codes it as part of Cundinamarca.
  if (p.adm1_code === "COL-1399") return { code: "CO-DC", name: "Bogotá" };
  return undefined;
}

function lookup<T>(table: Record<string, T>, key: Properties[string]): T {
  const value = table[String(key)];
  if (!value) throw new Error(`no entry for "${key}"`);
  return value;
}

/** Hashes everything that determines the output: this script, which pins the sources, and the mapshaper version. */
function inputStamp(): string {
  const require = createRequire(import.meta.url);
  const { version } = require("mapshaper/package.json") as { version: string };
  return createHash("sha256")
    .update(readFileSync(fileURLToPath(import.meta.url)))
    .update(`mapshaper@${version}`)
    .digest("hex");
}

function isUpToDate(stamp: string): boolean {
  if (!existsSync(MANIFEST_FILE)) return false;
  try {
    const manifest = JSON.parse(
      readFileSync(MANIFEST_FILE, "utf8"),
    ) as Manifest;
    return (
      manifest.stamp === stamp &&
      Object.values(manifest.files).every((file) =>
        existsSync(join(OUTPUT_DIR, file)),
      )
    );
  } catch {
    return false;
  }
}

async function fetchCollection(name: string): Promise<FeatureCollection> {
  const res = await fetch(`${NATURAL_EARTH_BASE}/${name}.geojson`, {
    signal: AbortSignal.timeout(DOWNLOAD_TIMEOUT_MS),
  });
  if (!res.ok) throw new Error(`failed to download ${name}: ${res.status}`);
  const body = Buffer.from(await res.arrayBuffer());
  const checksum = createHash("sha256").update(body).digest("hex");
  if (checksum !== NATURAL_EARTH_CHECKSUMS[name]) {
    throw new Error(
      `checksum mismatch for ${name}: expected ${NATURAL_EARTH_CHECKSUMS[name]}, got ${checksum}`,
    );
  }
  return JSON.parse(body.toString("utf8")) as FeatureCollection;
}

/** Runs the mapshaper commands on the collection and encodes it as TopoJSON with a single object named `layer`. */
async function toTopoJSON(
  layer: string,
  collection: FeatureCollection,
  commands: string[],
): Promise<string> {
  const output = await mapshaper.applyCommands(
    [
      "-i input.json",
      ...commands,
      `-rename-layers ${layer}`,
      "-o output.topo.json format=topojson id-field=id quantization=100000",
    ].join(" "),
    { "input.json": collection },
  );
  return String(output["output.topo.json"]);
}

/**
 * Replaces the output directory with the files and their manifest.
 * File names include a hash of the content, so browsers can cache them indefinitely.
 */
function writeOutput(stamp: string, files: Record<string, string>) {
  rmSync(OUTPUT_DIR, { recursive: true, force: true });
  mkdirSync(OUTPUT_DIR, { recursive: true });

  const manifest: Manifest = { stamp, files: {} };
  for (const [name, content] of Object.entries(files)) {
    const hash = createHash("sha256")
      .update(content)
      .digest("hex")
      .slice(0, 12);
    // "states/DE" is written as states-DE.<hash>.topo.json, keeping the output directory flat.
    const file = `${name.replace("/", "-")}.${hash}.topo.json`;
    writeFileSync(join(OUTPUT_DIR, file), content);
    manifest.files[name] = file;
  }
  // Written last, so an interrupted run leaves no manifest and is not mistaken for complete output.
  writeFileSync(MANIFEST_FILE, JSON.stringify(manifest, null, 2) + "\n");
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
