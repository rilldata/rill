import { describe, expect, it } from "vitest";
import type { V1Project } from "@rilldata/web-admin/client";
import {
  diffProjectTags,
  filterProjectsByTags,
  getProjectTags,
  isValidProjectTag,
  sanitizeProjectTag,
} from "./project-tags";

describe("sanitizeProjectTag", () => {
  it.each([
    ["Prod", "prod"],
    ["  stage ", "stage"],
    ["Data Science", "data-science"],
    ["env:prod", "env-prod"],
    ["team_a", "team_a"],
    ["--sandbox--", "sandbox"],
    ["Été", "été"],
    ["운영", "운영"],
    // Decomposed "é" (e + combining acute) is composed, as the admin service does.
    ["été", "été"],
    ["!!!", ""],
  ])("turns %j into %j", (input, expected) => {
    expect(sanitizeProjectTag(input)).toBe(expected);
  });
});

describe("isValidProjectTag", () => {
  it("accepts tags the admin service accepts", () => {
    expect(isValidProjectTag("prod")).toBe(true);
    expect(isValidProjectTag("eu-west-1")).toBe(true);
    expect(isValidProjectTag("운영")).toBe(true);
    expect(isValidProjectTag("a".repeat(40))).toBe(true);
  });

  it("rejects empty, long and malformed tags", () => {
    expect(isValidProjectTag("")).toBe(false);
    expect(isValidProjectTag("a".repeat(41))).toBe(false);
    expect(isValidProjectTag("_prod")).toBe(false);
    expect(isValidProjectTag("data science")).toBe(false);
  });
});

describe("diffProjectTags", () => {
  it("returns the tags to add and remove", () => {
    expect(
      diffProjectTags(["prod", "team-a"], ["stage", "team-a", "eu"]),
    ).toEqual({ addTags: ["stage", "eu"], removeTags: ["prod"] });
    expect(diffProjectTags(["prod"], ["prod"])).toEqual({
      addTags: [],
      removeTags: [],
    });
  });
});

describe("project tag filtering", () => {
  const projects: V1Project[] = [
    { name: "revenue", tags: ["prod", "eu"] },
    { name: "revenue-review", tags: ["sandbox"] },
    { name: "untagged" },
  ];

  it("lists the distinct tags of the projects", () => {
    expect(getProjectTags(projects)).toEqual(["eu", "prod", "sandbox"]);
    expect(getProjectTags([])).toEqual([]);
  });

  it("keeps projects with any of the selected tags", () => {
    const names = (tags: string[]) =>
      filterProjectsByTags(projects, tags).map((p) => p.name);
    expect(names([])).toEqual(["revenue", "revenue-review", "untagged"]);
    expect(names(["prod"])).toEqual(["revenue"]);
    expect(names(["eu", "sandbox"])).toEqual(["revenue", "revenue-review"]);
    expect(names(["missing"])).toEqual([]);
  });
});
