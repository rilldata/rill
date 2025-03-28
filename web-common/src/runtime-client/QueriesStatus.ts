import type { SupportedCompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { derived, writable } from "svelte/store";

type QueryStatusArg = {
  errorHeader: string;
  query: SupportedCompoundQueryResult<unknown, HTTPError>;
};
type ErrorPageProps = {
  statusCode: number | null | undefined;
  header: string;
  body: string;
  detail: string | undefined;
  fatal: boolean;
};

/**
 * Provides a consolidated loading and error status for queries.
 * Has additional `loadingForShortTime` and `loadingForLongTime` as extra information.
 * Also converts errors into params that can easily be fed into `ErrorPage`
 * TODO: is there a way to merge with getCompoundQuery?
 */
export class QueriesStatus {
  public readonly loading = writable(false);
  public readonly loadingForShortTime = writable(false);
  public readonly loadingForLongTime = writable(false);
  public readonly errors = writable<ErrorPageProps[]>([]);

  private readonly unsub: (() => void) | undefined;
  private loadingForShortTimeout: ReturnType<typeof setTimeout> | undefined;
  private loadingForLongTimeout: ReturnType<typeof setTimeout> | undefined;

  public constructor(
    private readonly queries: QueryStatusArg[],
    private readonly shortLoadingThreshold: number,
    private readonly longLoadingThreshold: number,
  ) {
    this.unsub = derived(
      queries.map(({ query }) => query),
      (queryResponses) => {
        const loading = queryResponses.some((q) => q.isLoading);
        const errors = queryResponses.map(
          (q) => q.error,
        ) as (HTTPError | null)[];

        return {
          loading,
          errors,
        };
      },
    ).subscribe(({ loading, errors }) => {
      if (loading) {
        this.setLoading();
      } else {
        this.resetLoading();
      }

      this.setErrors(errors);
    });
  }

  public teardown() {
    this.unsub?.();
  }

  private setLoading() {
    this.loading.set(true);

    if (!this.loadingForShortTimeout) {
      this.loadingForShortTimeout = setTimeout(() => {
        this.loadingForShortTime.set(true);
        this.loadingForShortTimeout = undefined;
      }, this.shortLoadingThreshold);
    }

    if (!this.loadingForLongTimeout) {
      this.loadingForLongTimeout = setTimeout(() => {
        this.loadingForLongTime.set(true);
        this.loadingForLongTimeout = undefined;
      }, this.longLoadingThreshold);
    }
  }

  private resetLoading() {
    this.loading.set(false);
    this.loadingForShortTime.set(false);
    this.loadingForLongTime.set(false);

    if (this.loadingForShortTimeout) clearTimeout(this.loadingForShortTimeout);
    this.loadingForShortTimeout = undefined;
    if (this.loadingForLongTimeout) clearTimeout(this.loadingForLongTimeout);
    this.loadingForLongTimeout = undefined;
  }

  private setErrors(errors: (HTTPError | null)[]) {
    const errorPageProps: ErrorPageProps[] = [];
    errors.forEach((e, i) => {
      if (!e) return;
      const message = e.response?.data?.message ?? e.message;
      errorPageProps.push({
        statusCode: e.response?.status,
        header: this.queries[i]?.errorHeader,
        body: "",
        detail: message,
        fatal: false,
      });
    });
    this.errors.set(errorPageProps);
  }
}
