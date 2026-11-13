/**
 * This file implements variants of the `CreateReport`, `EditReport`, `DeleteReport`, `TriggerReport` and `UnsubscribeReport` client code that authenticate using a bearer token.
 *
 * Modifications from the original Orval-generated code in `/web-admin/src/client/gen/default/default.ts` include:
 * - Request functions: Authentication via `Authorization: Bearer ${token}` header, replacing cookie-based authentication.
 * - Mutation creators: They delegate to the generated creators, overriding only `mutationFn`, so mutation keys and variables match the generated code.
 */

import type { QueryClient } from "@tanstack/svelte-query";
import {
  createAdminServiceCreateReport,
  createAdminServiceDeleteReport,
  createAdminServiceEditReport,
  createAdminServiceTriggerReport,
  createAdminServiceUnsubscribeReport,
  type AdminServiceCreateReportBodyBody,
  type AdminServiceTriggerReconcileBodyBody,
  type AdminServiceUnsubscribeAlertBodyBody,
  type V1CreateReportResponse,
  type V1DeleteReportResponse,
  type V1EditReportResponse,
  type V1TriggerReportResponse,
  type V1UnsubscribeReportResponse,
} from "@rilldata/web-admin/client";
import { httpClientWithToken } from "@rilldata/web-admin/client/http-client.ts";

export const adminServiceCreateReportUsingToken = (
  org: string,
  project: string,
  adminServiceCreateReportBodyBody: AdminServiceCreateReportBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1CreateReportResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/reports`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    data: adminServiceCreateReportBodyBody,
  });
};

export const createAdminServiceCreateReportUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceCreateReport(
    {
      mutation: {
        mutationFn: ({ org, project, data }) =>
          adminServiceCreateReportUsingToken(org, project, data, token),
      },
    },
    queryClient,
  );

export const adminServiceEditReportUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceCreateReportBodyBody: AdminServiceCreateReportBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1EditReportResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}`,
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    data: adminServiceCreateReportBodyBody,
  });
};

export const createAdminServiceEditReportUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceEditReport(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceEditReportUsingToken(org, project, name, data, token),
      },
    },
    queryClient,
  );

export const adminServiceDeleteReportUsingToken = (
  org: string,
  project: string,
  name: string,
  token: string,
) => {
  return httpClientWithToken<V1DeleteReportResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}`,
    method: "DELETE",
  });
};

export const createAdminServiceDeleteReportUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceDeleteReport(
    {
      mutation: {
        mutationFn: ({ org, project, name }) =>
          adminServiceDeleteReportUsingToken(org, project, name, token),
      },
    },
    queryClient,
  );

export const adminServiceTriggerReportUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceTriggerReconcileBodyBody: AdminServiceTriggerReconcileBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1TriggerReportResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}/trigger`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    data: adminServiceTriggerReconcileBodyBody,
  });
};

export const createAdminServiceTriggerReportUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceTriggerReport(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceTriggerReportUsingToken(org, project, name, data, token),
      },
    },
    queryClient,
  );

export const adminServiceUnsubscribeReportUsingToken = (
  org: string,
  project: string,
  name: string,
  adminServiceUnsubscribeAlertBodyBody: AdminServiceUnsubscribeAlertBodyBody,
  token: string,
) => {
  return httpClientWithToken<V1UnsubscribeReportResponse>(token, {
    url: `/v1/orgs/${org}/projects/${project}/reports/${name}/unsubscribe`,
    method: "POST",
    headers: { "Content-Type": "application/json" },
    data: adminServiceUnsubscribeAlertBodyBody,
  });
};

export const createAdminServiceUnsubscribeReportUsingToken = (
  token: string,
  queryClient?: QueryClient,
) =>
  createAdminServiceUnsubscribeReport(
    {
      mutation: {
        mutationFn: ({ org, project, name, data }) =>
          adminServiceUnsubscribeReportUsingToken(
            org,
            project,
            name,
            data,
            token,
          ),
      },
    },
    queryClient,
  );
