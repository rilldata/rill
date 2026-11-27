import { renderBannerMarkdown } from "@rilldata/web-common/components/banner/banner-markdown";
import DOMPurify from "dompurify";
import { describe, expect, it } from "vitest";

describe("renderBannerMarkdown", () => {
  it("Should leave plain text as it is", () => {
    expect(renderBannerMarkdown("Data refreshes daily at 6 AM UTC")).toBe(
      "Data refreshes daily at 6 AM UTC",
    );
    expect(renderBannerMarkdown("Rows with spend < 5 & clicks > 0")).toBe(
      "Rows with spend &lt; 5 &amp; clicks &gt; 0",
    );
  });

  it("Should render inline formatting", () => {
    expect(
      renderBannerMarkdown(
        "**Heads up:** the *March* numbers are ~~final~~ `preliminary`",
      ),
    ).toBe(
      "<strong>Heads up:</strong> the <em>March</em> numbers are <del>final</del> <code>preliminary</code>",
    );
  });

  it("Should open links in a new tab", () => {
    expect(
      renderBannerMarkdown("See the [status page](https://status.example.com)"),
    ).toBe(
      'See the <a href="https://status.example.com" target="_blank" rel="noopener noreferrer">status page</a>',
    );
    // Inline HTML links get the same treatment, whatever target they asked for.
    expect(renderBannerMarkdown('<a href="/docs" target="_top">Docs</a>')).toBe(
      '<a href="/docs" target="_blank" rel="noopener noreferrer">Docs</a>',
    );
  });

  it("Should keep the text of block-level markdown without rendering it", () => {
    expect(renderBannerMarkdown("# Maintenance on Friday")).toBe(
      "# Maintenance on Friday",
    );
  });

  it("Should show HTML other than inline formatting as written", () => {
    expect(renderBannerMarkdown("Use the <region> filter")).toBe(
      "Use the &lt;region&gt; filter",
    );

    const other =
      '<span style="position:fixed">styled</span> <img src="x" onerror="alert(1)"><script>alert(1)</script><style>body{display:none}</style>';
    const container = document.createElement("div");
    container.innerHTML = renderBannerMarkdown(
      `<u onclick="alert(1)">Note</u> ${other}`,
    );
    // The only element is the allowed tag, without its event handler.
    expect(
      [...container.querySelectorAll("*")].map((el) => el.outerHTML),
    ).toEqual(["<u>Note</u>"]);
    expect(container.textContent).toBe(`Note ${other}`);
  });

  it("Should remove unsafe links", () => {
    expect(renderBannerMarkdown("[click](javascript:alert(1))")).not.toContain(
      "javascript:",
    );
  });

  it("Should not change how the shared DOMPurify instance sanitizes", () => {
    renderBannerMarkdown("[link](https://example.com)");

    expect(DOMPurify.sanitize('<a href="https://example.com">link</a>')).toBe(
      '<a href="https://example.com">link</a>',
    );
  });
});
