import { writable } from "svelte/store";

// True only while a canvas PDF export is capturing. It gates the off-screen
// capture header (see CanvasPdfExportHeader) so the header exists in the DOM
// solely during export. Playwright locators (and the accessibility tree) match
// elements regardless of CSS visibility, so an always-mounted, merely-hidden
// header would still duplicate the live filter bar's text and labels; only
// removing it from the DOM avoids that.
export const canvasPdfExportActive = writable(false);
