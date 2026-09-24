import type { Page } from "@sveltejs/kit";

/**
 * Stand in for the `page` of `$app/state`, which only SvelteKit's client router populates.
 * Without it a component test reads SvelteKit's placeholder `a:` url, so code that compares its
 * state against `page.url` never sees a navigation.
 *
 * `PageMockForExploreTests` keeps this in sync with the `$app/stores` page it mocks, so a spec only
 * has to point the module at it:
 * ```
 * vi.mock("$app/state", async () => ({
 *   page: (
 *     await import(
 *       "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
 *     )
 *   ).pageStateMock,
 * }));
 * ```
 */
export const pageStateMock = $state({
  url: new URL("http://localhost/"),
  params: {},
  route: { id: null },
  data: {},
  state: {},
  status: 200,
  form: undefined,
} as unknown as Page);

/** Mirrors the `$app/stores` page onto the `$app/state` one. */
export function setPageStateMock(page: Page) {
  Object.assign(pageStateMock, page);
}
