import fs from "fs";
import path from "path";
import { describe, expect, it } from "vitest";
import en from "../messages/en.json";
import es from "../messages/es.json";

describe("i18n catalog integrity", () => {
  it("en and es have the same keys", () => {
    const enKeys = Object.keys(en).sort();
    const esKeys = Object.keys(es).sort();
    expect(enKeys).toEqual(esKeys);
  });

  it("no empty values in en", () => {
    const empty = Object.entries(en).filter(
      ([, v]) => typeof v === "string" && !v.trim(),
    );
    expect(empty).toEqual([]);
  });

  it("no empty values in es", () => {
    const empty = Object.entries(es).filter(
      ([, v]) => typeof v === "string" && !v.trim(),
    );
    expect(empty).toEqual([]);
  });

  it("parameters match between en and es", () => {
    const paramRe = /\{(\w+)\}/g;
    const mismatches: string[] = [];
    for (const key of Object.keys(en)) {
      const enVal = (en as Record<string, string>)[key] || "";
      const esVal = (es as Record<string, string>)[key] || "";
      const enParams = [...enVal.matchAll(paramRe)].map((m) => m[1]).sort();
      const esParams = [...esVal.matchAll(paramRe)].map((m) => m[1]).sort();
      if (JSON.stringify(enParams) !== JSON.stringify(esParams)) {
        mismatches.push(
          `${key}: en=${JSON.stringify(enParams)} es=${JSON.stringify(esParams)}`,
        );
      }
    }
    expect(mismatches).toEqual([]);
  });

  it("no duplicate keys in en.json", () => {
    const enRaw = fs.readFileSync(
      path.resolve(__dirname, "../messages/en.json"),
      "utf-8",
    );
    const keyMatches = enRaw.match(/"([^"]+)":/g) || [];
    const keys = keyMatches.map((k) => k.slice(1, -2));
    const uniqueKeys = new Set(keys);
    expect(keys.length).toBe(uniqueKeys.size);
  });

  it("no duplicate keys in es.json", () => {
    const esRaw = fs.readFileSync(
      path.resolve(__dirname, "../messages/es.json"),
      "utf-8",
    );
    const keyMatches = esRaw.match(/"([^"]+)":/g) || [];
    const keys = keyMatches.map((k) => k.slice(1, -2));
    const uniqueKeys = new Set(keys);
    expect(keys.length).toBe(uniqueKeys.size);
  });
});
