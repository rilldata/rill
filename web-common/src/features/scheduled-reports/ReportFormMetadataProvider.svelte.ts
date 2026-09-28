import { page } from "$app/state";
import {
  createAdminServiceCreateReport,
  createAdminServiceEditReport,
  createAdminServiceGetCurrentUser,
  createAdminServiceListProjectMemberUsers,
} from "@rilldata/web-admin/client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import {
  createRuntimeServiceGetInstance,
  createRuntimeServiceListNotifierConnectors,
} from "@rilldata/web-common/runtime-client";
import {
  createAdminServiceCreateReportUsingToken,
  createAdminServiceEditReportUsingToken,
} from "@rilldata/web-admin/features/scheduled-reports/report-client-using-token.ts";

export interface ReportFormMetadataProvider {
  mutation: ReturnType<
    typeof createAdminServiceCreateReport | typeof createAdminServiceEditReport
  >;

  organization: string;
  project: string;
  userEmail: string;
  projectMembersSet: Set<string>;
  hasSlackNotifier: boolean;

  cleanup(): void;
}

export class AdminReportFormMetadataProvider
  implements ReportFormMetadataProvider
{
  mutation: ReturnType<
    typeof createAdminServiceCreateReport | typeof createAdminServiceEditReport
  >;

  public organization: string;
  public project: string;
  public userEmail = $state<string>("");
  public projectMembersSet = $state<Set<string>>(new Set());
  public hasSlackNotifier = $state(false);

  private readonly userUnsub: () => void;
  private readonly projectMembersUnsub: () => void;
  private readonly notifiersUnsub: () => void;

  public constructor(client: RuntimeClient, isEditReport: boolean) {
    this.mutation = isEditReport
      ? createAdminServiceEditReport()
      : createAdminServiceCreateReport();

    this.organization = page.params.organization;
    this.project = page.params.project;

    const user = createAdminServiceGetCurrentUser();
    this.userUnsub = user.subscribe((userResp) => {
      if (!userResp.data?.user?.email) return;
      this.userEmail = userResp.data.user.email;
    });

    const listProjectMemberUsersQuery =
      createAdminServiceListProjectMemberUsers(this.organization, this.project);
    this.projectMembersUnsub = listProjectMemberUsersQuery.subscribe(
      (projectMembersResp) => {
        this.projectMembersSet = new Set(
          projectMembersResp.data?.members?.map((m) => m.userEmail!) ?? [],
        );
      },
    );

    const notifierConnectorsQuery = createRuntimeServiceListNotifierConnectors(
      client,
      {},
    );
    this.notifiersUnsub = notifierConnectorsQuery.subscribe(
      (notifierConnectorsResp) => {
        this.hasSlackNotifier =
          !!notifierConnectorsResp?.data?.connectors?.some(
            (c) => c.name === "slack",
          );
      },
    );
  }

  public cleanup() {
    this.userUnsub();
    this.projectMembersUnsub();
    this.notifiersUnsub();
  }
}

export class EmbedReportFormMetadataProvider
  implements ReportFormMetadataProvider
{
  mutation: ReturnType<
    typeof createAdminServiceCreateReport | typeof createAdminServiceEditReport
  >;

  public organization: string;
  public project: string;
  public userEmail = "";
  public projectMembersSet = new Set<string>();
  public hasSlackNotifier = false; // not support in embed

  private readonly instanceUnsub: () => void;

  public constructor(client: RuntimeClient, isEditReport: boolean) {
    const embedInstance = EmbedStore.getInstance();
    if (!embedInstance) {
      throw new Error("missing embed instance");
    }
    this.userEmail = embedInstance.userEmail;
    this.projectMembersSet = new Set([this.userEmail]);

    this.mutation = isEditReport
      ? createAdminServiceEditReportUsingToken(embedInstance.adminToken)
      : createAdminServiceCreateReportUsingToken(embedInstance.adminToken);

    const instanceQuery = createRuntimeServiceGetInstance(client, {});
    this.instanceUnsub = instanceQuery.subscribe((instanceResp) => {
      const url = instanceResp.data?.instance?.frontendUrl;
      if (!url) return { organization: "", project: "" };
      const parts = url.split("/");
      this.organization = parts[parts.length - 2] ?? "";
      this.project = parts[parts.length - 1] ?? "";
    });
  }

  public cleanup() {
    this.instanceUnsub();
  }
}
