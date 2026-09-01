import { describe, expect, it } from "vitest";
import { sanitizePageUrl } from "./sanitizePageUrl";

describe("sanitizePageUrl", () => {
  it("redacts the magic token in a public URL path", () => {
    expect(
      sanitizePageUrl(
        "https://ui.rilldata.com/myorg/myproject/-/share/rill_mgc_1yvUkhRrPE0VjwD8EbIluP0OOPz7gLm5/explore/revenue",
      ),
    ).toBe(
      "https://ui.rilldata.com/myorg/myproject/-/share/redacted/explore/revenue",
    );
  });

  it("redacts credential query params", () => {
    expect(
      sanitizePageUrl(
        "https://ui.rilldata.com/myorg/myproject/-/reports/weekly?token=rill_mgc_abc123",
      ),
    ).toBe(
      "https://ui.rilldata.com/myorg/myproject/-/reports/weekly?token=redacted",
    );

    expect(
      sanitizePageUrl("https://ui.rilldata.com/-/embed?access_token=eyJhbGc"),
    ).toBe("https://ui.rilldata.com/-/embed?access_token=redacted");
  });

  it("keeps dashboard state intact", () => {
    const url =
      "https://ui.rilldata.com/myorg/myproject/explore/revenue?view=pivot&tr=P7D&f=country+IN+%28%27US%27%29";
    expect(sanitizePageUrl(url)).toBe(url);
  });

  it("returns an empty string rather than passing through unparseable input", () => {
    expect(sanitizePageUrl("not a url")).toBe("");
  });
});
