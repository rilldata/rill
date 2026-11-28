import { describe, expect, it } from "vitest";
import { EmbedStore } from "./embed-store";

const REQUIRED_PARAMS =
  "instance_id=inst&runtime_host=https%3A%2F%2Fruntime.example.com&access_token=token";

function initEmbedStore(params: string) {
  EmbedStore.init(
    new URL(`https://ui.rilldata.com/-/embed?${REQUIRED_PARAMS}&${params}`),
  );
  return EmbedStore.getInstance()!;
}

describe("EmbedStore", () => {
  describe("navigationBarEnabled", () => {
    it("renders the navigation bar when navigation is enabled", () => {
      expect(initEmbedStore("navigation=true").navigationBarEnabled).toBe(true);
    });

    it("hides the navigation bar when hide_navigation_bar is set", () => {
      const embedStore = initEmbedStore(
        "navigation=true&hide_navigation_bar=true",
      );
      expect(embedStore.navigationBarEnabled).toBe(false);
      // Navigation itself stays enabled, so in-dashboard drill-through still works.
      expect(embedStore.navigationEnabled).toBe(true);
    });

    it("hides the navigation bar when navigation is disabled", () => {
      expect(initEmbedStore("navigation=false").navigationBarEnabled).toBe(
        false,
      );
    });

    it("hides the navigation bar when navigation is disabled, regardless of hide_navigation_bar", () => {
      expect(
        initEmbedStore("navigation=false&hide_navigation_bar=false")
          .navigationBarEnabled,
      ).toBe(false);
      expect(
        initEmbedStore("navigation=false&hide_navigation_bar=true")
          .navigationBarEnabled,
      ).toBe(false);
    });

    it("defaults to hidden when neither param is present", () => {
      const embedStore = initEmbedStore("");
      expect(embedStore.navigationEnabled).toBe(false);
      expect(embedStore.navigationBarEnabled).toBe(false);
    });
  });

  describe("lockedControls", () => {
    it("reads the locked_controls param as a map of control key to value", () => {
      const param = encodeURIComponent(
        JSON.stringify({ "deep_dive.publisher": "Acme,Globex%20Inc" }),
      );
      expect(initEmbedStore(`locked_controls=${param}`).lockedControls).toEqual(
        {
          "deep_dive.publisher": "Acme,Globex%20Inc",
        },
      );
    });

    it("ignores a missing or malformed param and non-string values", () => {
      expect(initEmbedStore("navigation=true").lockedControls).toEqual({});
      expect(initEmbedStore("locked_controls=nope").lockedControls).toEqual({});
      const param = encodeURIComponent(JSON.stringify({ a: 1, b: "x" }));
      expect(initEmbedStore(`locked_controls=${param}`).lockedControls).toEqual(
        {
          b: "x",
        },
      );
    });
  });
});
