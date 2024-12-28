# \AdminServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**admin_service_add_organization_member_user**](AdminServiceApi.md#admin_service_add_organization_member_user) | **POST** /v1/organizations/{organization}/members | AddOrganizationMemberUser adds a user to the organization
[**admin_service_add_organization_member_usergroup**](AdminServiceApi.md#admin_service_add_organization_member_usergroup) | **POST** /v1/organizations/{organization}/usergroups/{usergroup}/role | AddOrganizationMemberUsergroupRole adds the role for the user group
[**admin_service_add_project_member_user**](AdminServiceApi.md#admin_service_add_project_member_user) | **POST** /v1/organizations/{organization}/projects/{project}/members | AddProjectMemberUser adds a member to the project
[**admin_service_add_project_member_usergroup**](AdminServiceApi.md#admin_service_add_project_member_usergroup) | **POST** /v1/organizations/{organization}/projects/{project}/usergroups/{usergroup}/roles | AddProjectMemberUsergroupRole adds the role for the user group
[**admin_service_add_usergroup_member_user**](AdminServiceApi.md#admin_service_add_usergroup_member_user) | **PUT** /v1/organizations/{organization}/usergroups/{usergroup}/members/{email} | AddUsergroupMemberUser adds a member to the user group
[**admin_service_approve_project_access**](AdminServiceApi.md#admin_service_approve_project_access) | **POST** /v1/project-access-request/{id}/approve | 
[**admin_service_cancel_billing_subscription**](AdminServiceApi.md#admin_service_cancel_billing_subscription) | **DELETE** /v1/organizations/{organization}/billing/subscriptions | CancelBillingSubscription cancels the billing subscription for the organization
[**admin_service_connect_project_to_github**](AdminServiceApi.md#admin_service_connect_project_to_github) | **POST** /v1/organizations/{organization}/projects/{project}/connect-to-github | Connects a rill managed project to github. Replaces the contents of the remote repo with the contents of the project.
[**admin_service_create_alert**](AdminServiceApi.md#admin_service_create_alert) | **POST** /v1/organizations/{organization}/projects/{project}/alerts | CreateAlert adds a virtual file for an alert, triggers a reconcile, and waits for the alert to be added to the runtime catalog
[**admin_service_create_asset**](AdminServiceApi.md#admin_service_create_asset) | **POST** /v1/organizations/{organizationName}/create_asset | CreateAsset returns a one time signed URL using which any asset can be uploaded.
[**admin_service_create_bookmark**](AdminServiceApi.md#admin_service_create_bookmark) | **POST** /v1/users/bookmarks | CreateBookmark creates a bookmark for the given user or for all users for the dashboard
[**admin_service_create_organization**](AdminServiceApi.md#admin_service_create_organization) | **POST** /v1/organizations | CreateOrganization creates a new organization
[**admin_service_create_project**](AdminServiceApi.md#admin_service_create_project) | **POST** /v1/organizations/{organizationName}/projects | CreateProject creates a new project
[**admin_service_create_project_whitelisted_domain**](AdminServiceApi.md#admin_service_create_project_whitelisted_domain) | **POST** /v1/organizations/{organization}/projects/{project}/whitelisted | CreateProjectWhitelistedDomain adds a domain to the project's whitelisted
[**admin_service_create_report**](AdminServiceApi.md#admin_service_create_report) | **POST** /v1/organizations/{organization}/projects/{project}/reports | CreateReport adds a virtual file for a report, triggers a reconcile, and waits for the report to be added to the runtime catalog
[**admin_service_create_service**](AdminServiceApi.md#admin_service_create_service) | **POST** /v1/organizations/{organizationName}/services | CreateService creates a new service per organization
[**admin_service_create_usergroup**](AdminServiceApi.md#admin_service_create_usergroup) | **POST** /v1/organizations/{organization}/usergroups | CreateUsergroup creates a user group in the organization
[**admin_service_create_whitelisted_domain**](AdminServiceApi.md#admin_service_create_whitelisted_domain) | **POST** /v1/organizations/{organization}/whitelisted | CreateWhitelistedDomain adds a domain to the whitelist
[**admin_service_delete_alert**](AdminServiceApi.md#admin_service_delete_alert) | **DELETE** /v1/organizations/{organization}/projects/{project}/alerts/{name} | DeleteAlert deletes the virtual file for a UI-managed alert, triggers a reconcile, and waits for the alert to be deleted in the runtime
[**admin_service_delete_organization**](AdminServiceApi.md#admin_service_delete_organization) | **DELETE** /v1/organizations/{name} | DeleteOrganization deletes an organizations
[**admin_service_delete_project**](AdminServiceApi.md#admin_service_delete_project) | **DELETE** /v1/organizations/{organizationName}/projects/{name} | DeleteProject deletes an project
[**admin_service_delete_report**](AdminServiceApi.md#admin_service_delete_report) | **DELETE** /v1/organizations/{organization}/projects/{project}/reports/{name} | DeleteReport deletes the virtual file for a UI-managed report, triggers a reconcile, and waits for the report to be deleted in the runtime
[**admin_service_delete_service**](AdminServiceApi.md#admin_service_delete_service) | **DELETE** /v1/organizations/{organizationName}/services/{name} | DeleteService deletes a service per organization
[**admin_service_delete_usergroup**](AdminServiceApi.md#admin_service_delete_usergroup) | **DELETE** /v1/organizations/{organization}/usergroups/{usergroup} | DeleteUsergroup deletes the user group from the organization
[**admin_service_deny_project_access**](AdminServiceApi.md#admin_service_deny_project_access) | **POST** /v1/project-access-request/{id}/deny | 
[**admin_service_edit_alert**](AdminServiceApi.md#admin_service_edit_alert) | **PUT** /v1/organizations/{organization}/projects/{project}/alerts/{name} | EditAlert edits a virtual file for a UI-managed alert, triggers a reconcile, and waits for the alert to be updated in the runtime
[**admin_service_edit_report**](AdminServiceApi.md#admin_service_edit_report) | **PUT** /v1/organizations/{organization}/projects/{project}/reports/{name} | EditReport edits a virtual file for a UI-managed report, triggers a reconcile, and waits for the report to be updated in the runtime
[**admin_service_edit_usergroup**](AdminServiceApi.md#admin_service_edit_usergroup) | **PUT** /v1/organizations/{organization}/usergroups/{usergroup} | EditUsergroup renames the user group
[**admin_service_generate_alert_yaml**](AdminServiceApi.md#admin_service_generate_alert_yaml) | **POST** /v1/organizations/{organization}/projects/{project}/alerts/-/yaml | GenerateAlertYAML generates YAML for an alert to be copied into a project's Git repository
[**admin_service_generate_report_yaml**](AdminServiceApi.md#admin_service_generate_report_yaml) | **POST** /v1/organizations/{organization}/projects/{project}/reports/-/yaml | GenerateReportYAML generates YAML for a scheduled report to be copied into a project's Git repository
[**admin_service_get_alert_meta**](AdminServiceApi.md#admin_service_get_alert_meta) | **POST** /v1/projects/{projectId}/alerts/meta | GetAlertMeta returns metadata for checking an alert. It's currently only called by the alert reconciler in the runtime.
[**admin_service_get_alert_yaml**](AdminServiceApi.md#admin_service_get_alert_yaml) | **GET** /v1/organizations/{organization}/projects/{project}/alerts/{name}/yaml | GenerateAlertYAML generates YAML for an alert to be copied into a project's Git repository
[**admin_service_get_billing_project_credentials**](AdminServiceApi.md#admin_service_get_billing_project_credentials) | **POST** /v1/billing/metrics-project-credentials | GetBillingProjectCredentials returns credentials for the configured cloud metrics project filtered by request organization
[**admin_service_get_billing_subscription**](AdminServiceApi.md#admin_service_get_billing_subscription) | **GET** /v1/organizations/{organization}/billing/subscriptions | GetBillingSubscription lists the subscription for the organization
[**admin_service_get_bookmark**](AdminServiceApi.md#admin_service_get_bookmark) | **GET** /v1/users/bookmarks/{bookmarkId} | GetBookmark returns the bookmark for the given user for the given project
[**admin_service_get_clone_credentials**](AdminServiceApi.md#admin_service_get_clone_credentials) | **GET** /v1/organizations/{organization}/projects/{project}/clone-credentials | GetCloneCredentials returns credentials and other details for a project's Git repository or archive path if git repo is not configured.
[**admin_service_get_current_magic_auth_token**](AdminServiceApi.md#admin_service_get_current_magic_auth_token) | **GET** /v1/magic-tokens/current | GetCurrentMagicAuthToken returns information about the current magic auth token.
[**admin_service_get_current_user**](AdminServiceApi.md#admin_service_get_current_user) | **GET** /v1/users/current | GetCurrentUser returns the currently authenticated user (if any)
[**admin_service_get_deployment_credentials**](AdminServiceApi.md#admin_service_get_deployment_credentials) | **POST** /v1/organizations/{organization}/projects/{project}/credentials | GetDeploymentCredentials returns runtime info and access token on behalf of a specific user, or alternatively for a raw set of JWT attributes
[**admin_service_get_github_repo_status**](AdminServiceApi.md#admin_service_get_github_repo_status) | **GET** /v1/github/repositories | GetGithubRepoRequest returns info about a Github repo based on the caller's installations. If the caller has not granted access to the repository, instructions for granting access are returned.
[**admin_service_get_github_user_status**](AdminServiceApi.md#admin_service_get_github_user_status) | **GET** /v1/github/user | GetGithubUserStatus returns info about a Github user account based on the caller's installations. If we don't have access to user's personal account tokens or it is expired, instructions for granting access are returned.
[**admin_service_get_i_frame**](AdminServiceApi.md#admin_service_get_i_frame) | **POST** /v1/organizations/{organization}/projects/{project}/iframe | GetIFrame returns the iframe URL for the given project
[**admin_service_get_organization**](AdminServiceApi.md#admin_service_get_organization) | **GET** /v1/organizations/{name} | GetOrganization returns information about a specific organization
[**admin_service_get_organization_name_for_domain**](AdminServiceApi.md#admin_service_get_organization_name_for_domain) | **GET** /v1/organization-for-domain/{domain} | GetOrganizationNameForDomain finds the org name for a custom domain. If the application detects it is running on a non-default domain, it can use this to find the org to present. It can be called without being authenticated.
[**admin_service_get_payments_portal_url**](AdminServiceApi.md#admin_service_get_payments_portal_url) | **GET** /v1/organizations/{organization}/billing/payments/portal-url | GetPaymentsPortalURL returns the URL for the billing session to collect payment method
[**admin_service_get_project**](AdminServiceApi.md#admin_service_get_project) | **GET** /v1/organizations/{organizationName}/projects/{name} | GetProject returns information about a specific project
[**admin_service_get_project_access_request**](AdminServiceApi.md#admin_service_get_project_access_request) | **GET** /v1/project-access-request/{id} | 
[**admin_service_get_project_by_id**](AdminServiceApi.md#admin_service_get_project_by_id) | **GET** /v1/projects/{id} | GetProject returns information about a specific project
[**admin_service_get_project_variables**](AdminServiceApi.md#admin_service_get_project_variables) | **GET** /v1/organizations/{organization}/projects/{project}/variables | GetProjectVariables returns project variables.
[**admin_service_get_repo_meta**](AdminServiceApi.md#admin_service_get_repo_meta) | **GET** /v1/projects/{projectId}/repo/meta | GetRepoMeta returns credentials and other metadata for accessing a project's repo
[**admin_service_get_report_meta**](AdminServiceApi.md#admin_service_get_report_meta) | **POST** /v1/projects/{projectId}/reports/meta | GetReportMeta returns metadata for generating a report. It's currently only called by the report reconciler in the runtime.
[**admin_service_get_user**](AdminServiceApi.md#admin_service_get_user) | **GET** /v1/users | GetUser returns user by email
[**admin_service_get_usergroup**](AdminServiceApi.md#admin_service_get_usergroup) | **GET** /v1/organizations/{organization}/usergroups/{usergroup} | GetUsergroups returns the user group details
[**admin_service_hibernate_project**](AdminServiceApi.md#admin_service_hibernate_project) | **POST** /v1/organizations/{organization}/projects/{project}/hibernate | HibernateProject hibernates a project by tearing down its deployments.
[**admin_service_issue_magic_auth_token**](AdminServiceApi.md#admin_service_issue_magic_auth_token) | **POST** /v1/organizations/{organization}/projects/{project}/tokens/magic | IssueMagicAuthToken creates a \"magic\" auth token that provides limited access to a specific filtered dashboard in a specific project.
[**admin_service_issue_representative_auth_token**](AdminServiceApi.md#admin_service_issue_representative_auth_token) | **POST** /v1/tokens/represent | IssueRepresentativeAuthToken returns the temporary token for given email
[**admin_service_issue_service_auth_token**](AdminServiceApi.md#admin_service_issue_service_auth_token) | **POST** /v1/organizations/{organizationName}/services/{serviceName}/tokens | IssueServiceAuthToken returns the temporary token for given service account
[**admin_service_leave_organization**](AdminServiceApi.md#admin_service_leave_organization) | **DELETE** /v1/organizations/{organization}/members/current | LeaveOrganization removes the current user from the organization
[**admin_service_list_bookmarks**](AdminServiceApi.md#admin_service_list_bookmarks) | **GET** /v1/users/bookmarks | ListBookmarks lists all the bookmarks for the user and global ones for dashboard
[**admin_service_list_github_user_repos**](AdminServiceApi.md#admin_service_list_github_user_repos) | **GET** /v1/github/user/repositories | 
[**admin_service_list_magic_auth_tokens**](AdminServiceApi.md#admin_service_list_magic_auth_tokens) | **GET** /v1/organizations/{organization}/projects/{project}/tokens/magic | ListMagicAuthTokens lists all the magic auth tokens for a specific project.
[**admin_service_list_organization_billing_issues**](AdminServiceApi.md#admin_service_list_organization_billing_issues) | **GET** /v1/organizations/{organization}/billing/issues | ListOrganizationBillingIssues lists all the billing issues for the organization
[**admin_service_list_organization_invites**](AdminServiceApi.md#admin_service_list_organization_invites) | **GET** /v1/organizations/{organization}/invites | ListOrganizationInvites lists all the org invites
[**admin_service_list_organization_member_usergroups**](AdminServiceApi.md#admin_service_list_organization_member_usergroups) | **GET** /v1/organizations/{organization}/usergroups | ListOrganizationMemberUsergroups lists the org's user groups
[**admin_service_list_organization_member_users**](AdminServiceApi.md#admin_service_list_organization_member_users) | **GET** /v1/organizations/{organization}/members | ListOrganizationMemberUsers lists all the org members
[**admin_service_list_organizations**](AdminServiceApi.md#admin_service_list_organizations) | **GET** /v1/organizations | ListOrganizations lists all the organizations currently managed by the admin
[**admin_service_list_project_invites**](AdminServiceApi.md#admin_service_list_project_invites) | **GET** /v1/organizations/{organization}/projects/{project}/invites | ListProjectInvites lists all the project invites
[**admin_service_list_project_member_usergroups**](AdminServiceApi.md#admin_service_list_project_member_usergroups) | **GET** /v1/organizations/{organization}/project/{project}/usergroups | ListProjectMemberUsergroups lists the org's user groups
[**admin_service_list_project_member_users**](AdminServiceApi.md#admin_service_list_project_member_users) | **GET** /v1/organizations/{organization}/projects/{project}/members | ListProjectMemberUsers lists all the project members
[**admin_service_list_project_whitelisted_domains**](AdminServiceApi.md#admin_service_list_project_whitelisted_domains) | **GET** /v1/organizations/{organization}/projects/{project}/whitelisted | ListWhitelistedDomains lists all the whitelisted domains of the project
[**admin_service_list_projects_for_organization**](AdminServiceApi.md#admin_service_list_projects_for_organization) | **GET** /v1/organizations/{organizationName}/projects | ListProjectsForOrganization lists all the projects currently available for given organizations
[**admin_service_list_public_billing_plans**](AdminServiceApi.md#admin_service_list_public_billing_plans) | **GET** /v1/billing/plans | ListPublicBillingPlans lists all public billing plans
[**admin_service_list_service_auth_tokens**](AdminServiceApi.md#admin_service_list_service_auth_tokens) | **GET** /v1/organizations/{organizationName}/services/{serviceName}/tokens | ListServiceAuthTokens lists all the service auth tokens
[**admin_service_list_services**](AdminServiceApi.md#admin_service_list_services) | **GET** /v1/organizations/{organizationName}/services | ListService returns all the services per organization
[**admin_service_list_superusers**](AdminServiceApi.md#admin_service_list_superusers) | **GET** /v1/superuser/members | ListSuperusers lists all the superusers
[**admin_service_list_usergroup_member_users**](AdminServiceApi.md#admin_service_list_usergroup_member_users) | **GET** /v1/organizations/{organization}/usergroups/{usergroup}/members | ListUsergroupMemberUsers lists all the user group members
[**admin_service_list_whitelisted_domains**](AdminServiceApi.md#admin_service_list_whitelisted_domains) | **GET** /v1/organizations/{organization}/whitelisted | ListWhitelistedDomains lists all the whitelisted domains for the organization
[**admin_service_ping**](AdminServiceApi.md#admin_service_ping) | **GET** /v1/ping | Ping returns information about the server
[**admin_service_provision**](AdminServiceApi.md#admin_service_provision) | **POST** /v1/deployments/{deploymentId}/provision | Provision provisions a new resource for a deployment. If an existing resource matches the request, it will be returned without provisioning a new resource.
[**admin_service_pull_virtual_repo**](AdminServiceApi.md#admin_service_pull_virtual_repo) | **GET** /v1/projects/{projectId}/repo/virtual | PullVirtualRepo fetches files from a project's virtual repo
[**admin_service_redeploy_project**](AdminServiceApi.md#admin_service_redeploy_project) | **POST** /v1/organizations/{organization}/projects/{project}/redeploy | RedeployProject creates a new production deployment for a project. If the project currently has another production deployment, the old deployment will be deprovisioned. This RPC can be used to redeploy a project that has been hibernated.
[**admin_service_remove_bookmark**](AdminServiceApi.md#admin_service_remove_bookmark) | **DELETE** /v1/users/bookmarks/{bookmarkId} | RemoveBookmark removes the bookmark for the given user or all users
[**admin_service_remove_organization_member_user**](AdminServiceApi.md#admin_service_remove_organization_member_user) | **DELETE** /v1/organizations/{organization}/members/{email} | RemoveOrganizationMemberUser removes member from the organization
[**admin_service_remove_organization_member_usergroup**](AdminServiceApi.md#admin_service_remove_organization_member_usergroup) | **DELETE** /v1/organizations/{organization}/usergroups/{usergroup}/role | RemoveOrganizationMemberUsergroup revokes the organization-level role for the user group
[**admin_service_remove_project_member_user**](AdminServiceApi.md#admin_service_remove_project_member_user) | **DELETE** /v1/organizations/{organization}/projects/{project}/members/{email} | RemoveProjectMemberUser removes member from the project
[**admin_service_remove_project_member_usergroup**](AdminServiceApi.md#admin_service_remove_project_member_usergroup) | **DELETE** /v1/organizations/{organization}/projects/{project}/usergroups/{usergroup}/roles | RemoveProjectMemberUsergroup revokes the project-level role for the user group
[**admin_service_remove_project_whitelisted_domain**](AdminServiceApi.md#admin_service_remove_project_whitelisted_domain) | **DELETE** /v1/organizations/{organization}/projects/{project}/whitelisted/{domain} | RemoveProjectWhitelistedDomain removes a domain from the project's whitelisted
[**admin_service_remove_usergroup_member_user**](AdminServiceApi.md#admin_service_remove_usergroup_member_user) | **DELETE** /v1/organizations/{organization}/usergroups/{usergroup}/members/{email} | RemoveUsergroupMemberUser removes member from the user group
[**admin_service_remove_whitelisted_domain**](AdminServiceApi.md#admin_service_remove_whitelisted_domain) | **DELETE** /v1/organizations/{organization}/whitelisted/{domain} | RemoveWhitelistedDomain removes a domain from the whitelist list
[**admin_service_rename_usergroup**](AdminServiceApi.md#admin_service_rename_usergroup) | **POST** /v1/organizations/{organization}/usergroups/{usergroup} | RenameUsergroup renames the user group
[**admin_service_renew_billing_subscription**](AdminServiceApi.md#admin_service_renew_billing_subscription) | **POST** /v1/organizations/{organization}/billing/subscriptions/renew | RenewBillingSubscription renews the billing plan for the organization once cancelled
[**admin_service_request_project_access**](AdminServiceApi.md#admin_service_request_project_access) | **POST** /v1/organizations/{organization}/projects/{project}/request-access | 
[**admin_service_revoke_current_auth_token**](AdminServiceApi.md#admin_service_revoke_current_auth_token) | **DELETE** /v1/tokens/current | RevokeCurrentAuthToken revoke the current auth token
[**admin_service_revoke_magic_auth_token**](AdminServiceApi.md#admin_service_revoke_magic_auth_token) | **DELETE** /v1/magic-tokens/{tokenId} | RevokeMagicAuthToken revokes a magic auth token.
[**admin_service_revoke_service_auth_token**](AdminServiceApi.md#admin_service_revoke_service_auth_token) | **DELETE** /v1/services/tokens/{tokenId} | RevokeServiceAuthToken revoke the service auth token
[**admin_service_search_project_names**](AdminServiceApi.md#admin_service_search_project_names) | **GET** /v1/superuser/projects/search | SearchProjectNames returns project names matching the pattern
[**admin_service_search_project_users**](AdminServiceApi.md#admin_service_search_project_users) | **GET** /v1/organizations/{organization}/projects/{project}/users/search | SearchProjectUsers returns users who has access to to a project (including org members that have access through a usergroup)
[**admin_service_search_users**](AdminServiceApi.md#admin_service_search_users) | **GET** /v1/users/search | GetUsersByEmail returns users by email
[**admin_service_set_organization_member_user_role**](AdminServiceApi.md#admin_service_set_organization_member_user_role) | **PUT** /v1/organizations/{organization}/members/{email} | SetOrganizationMemberUserRole sets the role for the member
[**admin_service_set_organization_member_usergroup_role**](AdminServiceApi.md#admin_service_set_organization_member_usergroup_role) | **PUT** /v1/organizations/{organization}/usergroups/{usergroup}/role | SetOrganizationMemberUsergroupRole sets the role for the user group
[**admin_service_set_project_member_user_role**](AdminServiceApi.md#admin_service_set_project_member_user_role) | **PUT** /v1/organizations/{organization}/projects/{project}/members/{email} | SetProjectMemberUserRole sets the role for the member
[**admin_service_set_project_member_usergroup_role**](AdminServiceApi.md#admin_service_set_project_member_usergroup_role) | **PUT** /v1/organizations/{organization}/projects/{project}/usergroups/{usergroup}/roles | SetProjectMemberUsergroupRole sets the role for the user group
[**admin_service_set_superuser**](AdminServiceApi.md#admin_service_set_superuser) | **POST** /v1/superuser/members | SetSuperuser adds/remove a superuser
[**admin_service_sudo_delete_organization_billing_issue**](AdminServiceApi.md#admin_service_sudo_delete_organization_billing_issue) | **DELETE** /v1/superuser/organizations/{organization}/billing/issues/{type} | SudoDeleteOrganizationBillingIssue deletes a billing issue of a type for the organization
[**admin_service_sudo_extend_trial**](AdminServiceApi.md#admin_service_sudo_extend_trial) | **POST** /v1/superuser/organization/trial/extend | SudoExtendTrial extends the trial period for an organization
[**admin_service_sudo_get_resource**](AdminServiceApi.md#admin_service_sudo_get_resource) | **GET** /v1/superuser/resource | SudoGetResource returns details about a resource by ID lookup
[**admin_service_sudo_issue_runtime_manager_token**](AdminServiceApi.md#admin_service_sudo_issue_runtime_manager_token) | **POST** /v1/superuser/deployments/manager-token | SudoIssueRuntimeManagerToken returns a runtime JWT with full manager permissions for a runtime.
[**admin_service_sudo_trigger_billing_repair**](AdminServiceApi.md#admin_service_sudo_trigger_billing_repair) | **POST** /v1/superuser/billing/repair | SudoTriggerBillingRepair triggers billing repair jobs for orgs that doesn't have billing info and puts them on trial
[**admin_service_sudo_update_annotations**](AdminServiceApi.md#admin_service_sudo_update_annotations) | **PATCH** /v1/superuser/projects/annotations | SudoUpdateAnnotations endpoint for superusers to update project annotations
[**admin_service_sudo_update_organization_billing_customer**](AdminServiceApi.md#admin_service_sudo_update_organization_billing_customer) | **PATCH** /v1/superuser/organization/billing/customer_id | SudoUpdateOrganizationBillingCustomer update the billing customer for the organization
[**admin_service_sudo_update_organization_custom_domain**](AdminServiceApi.md#admin_service_sudo_update_organization_custom_domain) | **PATCH** /v1/superuser/organization/custom-domain | SudoUpdateOrganizationCustomDomain updates the custom domain for an organization. It only updates the custom domain in the database, which is used to ensure correct redirects. The DNS records and ingress TLS must be configured separately.
[**admin_service_sudo_update_organization_quotas**](AdminServiceApi.md#admin_service_sudo_update_organization_quotas) | **PATCH** /v1/superuser/quotas/organization | SudoUpdateOrganizationQuotas update the quotas available for orgs
[**admin_service_sudo_update_user_quotas**](AdminServiceApi.md#admin_service_sudo_update_user_quotas) | **PATCH** /v1/superuser/quotas/user | SudoUpdateUserQuotas update the quotas for users
[**admin_service_trigger_reconcile**](AdminServiceApi.md#admin_service_trigger_reconcile) | **POST** /v1/deployments/{deploymentId}/reconcile | TriggerReconcile triggers reconcile for the project's prod deployment. DEPRECATED: Clients should call CreateTrigger directly on the deployed runtime instead.
[**admin_service_trigger_redeploy**](AdminServiceApi.md#admin_service_trigger_redeploy) | **POST** /v1/projects/-/redeploy | TriggerRedeploy is similar to RedeployProject. DEPRECATED: Use RedeployProject instead.
[**admin_service_trigger_refresh_sources**](AdminServiceApi.md#admin_service_trigger_refresh_sources) | **POST** /v1/deployments/{deploymentId}/refresh | TriggerRefreshSources refresh the source for production deployment. DEPRECATED: Clients should call CreateTrigger directly on the deployed runtime instead.
[**admin_service_trigger_report**](AdminServiceApi.md#admin_service_trigger_report) | **POST** /v1/organizations/{organization}/projects/{project}/reports/{name}/trigger | TriggerReport triggers an ad-hoc report run
[**admin_service_unsubscribe_alert**](AdminServiceApi.md#admin_service_unsubscribe_alert) | **POST** /v1/organizations/{organization}/projects/{project}/alerts/{name}/unsubscribe | UnsubscribeAlert removes the calling user from a alert's recipients list
[**admin_service_unsubscribe_report**](AdminServiceApi.md#admin_service_unsubscribe_report) | **POST** /v1/organizations/{organization}/projects/{project}/reports/{name}/unsubscribe | UnsubscribeReport removes the calling user from a reports recipients list
[**admin_service_update_billing_subscription**](AdminServiceApi.md#admin_service_update_billing_subscription) | **PATCH** /v1/organizations/{organization}/billing/subscriptions | UpdateBillingSubscription updates the billing plan for the organization
[**admin_service_update_bookmark**](AdminServiceApi.md#admin_service_update_bookmark) | **PUT** /v1/users/bookmarks | UpdateBookmark updates a bookmark for the given user for the given project
[**admin_service_update_organization**](AdminServiceApi.md#admin_service_update_organization) | **PATCH** /v1/organizations/{name} | UpdateOrganization deletes an organizations
[**admin_service_update_project**](AdminServiceApi.md#admin_service_update_project) | **PATCH** /v1/organizations/{organizationName}/projects/{name} | UpdateProject updates a project
[**admin_service_update_project_variables**](AdminServiceApi.md#admin_service_update_project_variables) | **PUT** /v1/organizations/{organization}/projects/{project}/variables | UpdateProjectVariables updates variables for a project.
[**admin_service_update_service**](AdminServiceApi.md#admin_service_update_service) | **PATCH** /v1/organizations/{organizationName}/services/{name} | UpdateService updates a service per organization
[**admin_service_update_user_preferences**](AdminServiceApi.md#admin_service_update_user_preferences) | **PUT** /v1/users/preferences | UpdateUserPreferences updates the preferences for the user
[**admin_service_upload_project_assets**](AdminServiceApi.md#admin_service_upload_project_assets) | **POST** /v1/organizations/{organization}/projects/{project}/upload-assets | Converts a project connected to github to a rill managed project. Uploads the current project to assets.



## admin_service_add_organization_member_user

> models::V1AddOrganizationMemberUserResponse admin_service_add_organization_member_user(organization, body)
AddOrganizationMemberUser adds a user to the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**body** | [**AdminServiceAddOrganizationMemberUserRequest**](AdminServiceAddOrganizationMemberUserRequest.md) |  | [required] |

### Return type

[**models::V1AddOrganizationMemberUserResponse**](v1AddOrganizationMemberUserResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_add_organization_member_usergroup

> serde_json::Value admin_service_add_organization_member_usergroup(organization, usergroup, body)
AddOrganizationMemberUsergroupRole adds the role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_add_project_member_user

> models::V1AddProjectMemberUserResponse admin_service_add_project_member_user(organization, project, body)
AddProjectMemberUser adds a member to the project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceAddProjectMemberUserRequest**](AdminServiceAddProjectMemberUserRequest.md) |  | [required] |

### Return type

[**models::V1AddProjectMemberUserResponse**](v1AddProjectMemberUserResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_add_project_member_usergroup

> serde_json::Value admin_service_add_project_member_usergroup(organization, project, usergroup, body)
AddProjectMemberUsergroupRole adds the role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_add_usergroup_member_user

> serde_json::Value admin_service_add_usergroup_member_user(organization, usergroup, email, body)
AddUsergroupMemberUser adds a member to the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**email** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_approve_project_access

> serde_json::Value admin_service_approve_project_access(id, body)


### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**id** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_cancel_billing_subscription

> serde_json::Value admin_service_cancel_billing_subscription(organization)
CancelBillingSubscription cancels the billing subscription for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_connect_project_to_github

> serde_json::Value admin_service_connect_project_to_github(organization, project, body)
Connects a rill managed project to github. Replaces the contents of the remote repo with the contents of the project.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceConnectProjectToGithubRequest**](AdminServiceConnectProjectToGithubRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_alert

> models::V1CreateAlertResponse admin_service_create_alert(organization, project, body)
CreateAlert adds a virtual file for an alert, triggers a reconcile, and waits for the alert to be added to the runtime catalog

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceCreateAlertRequest**](AdminServiceCreateAlertRequest.md) |  | [required] |

### Return type

[**models::V1CreateAlertResponse**](v1CreateAlertResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_asset

> models::V1CreateAssetResponse admin_service_create_asset(organization_name, body)
CreateAsset returns a one time signed URL using which any asset can be uploaded.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**body** | [**AdminServiceCreateAssetRequest**](AdminServiceCreateAssetRequest.md) |  | [required] |

### Return type

[**models::V1CreateAssetResponse**](v1CreateAssetResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_bookmark

> models::V1CreateBookmarkResponse admin_service_create_bookmark(body)
CreateBookmark creates a bookmark for the given user or for all users for the dashboard

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1CreateBookmarkRequest**](V1CreateBookmarkRequest.md) |  | [required] |

### Return type

[**models::V1CreateBookmarkResponse**](v1CreateBookmarkResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_organization

> models::V1CreateOrganizationResponse admin_service_create_organization(body)
CreateOrganization creates a new organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1CreateOrganizationRequest**](V1CreateOrganizationRequest.md) |  | [required] |

### Return type

[**models::V1CreateOrganizationResponse**](v1CreateOrganizationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_project

> models::V1CreateProjectResponse admin_service_create_project(organization_name, body)
CreateProject creates a new project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**body** | [**AdminServiceCreateProjectRequest**](AdminServiceCreateProjectRequest.md) |  | [required] |

### Return type

[**models::V1CreateProjectResponse**](v1CreateProjectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_project_whitelisted_domain

> serde_json::Value admin_service_create_project_whitelisted_domain(organization, project, body)
CreateProjectWhitelistedDomain adds a domain to the project's whitelisted

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceCreateProjectWhitelistedDomainRequest**](AdminServiceCreateProjectWhitelistedDomainRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_report

> models::V1CreateReportResponse admin_service_create_report(organization, project, body)
CreateReport adds a virtual file for a report, triggers a reconcile, and waits for the report to be added to the runtime catalog

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceCreateReportRequest**](AdminServiceCreateReportRequest.md) |  | [required] |

### Return type

[**models::V1CreateReportResponse**](v1CreateReportResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_service

> models::V1CreateServiceResponse admin_service_create_service(organization_name, name)
CreateService creates a new service per organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | Option<**String**> |  |  |

### Return type

[**models::V1CreateServiceResponse**](v1CreateServiceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_usergroup

> serde_json::Value admin_service_create_usergroup(organization, body)
CreateUsergroup creates a user group in the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**body** | [**AdminServiceCreateUsergroupRequest**](AdminServiceCreateUsergroupRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_create_whitelisted_domain

> serde_json::Value admin_service_create_whitelisted_domain(organization, body)
CreateWhitelistedDomain adds a domain to the whitelist

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**body** | [**AdminServiceCreateProjectWhitelistedDomainRequest**](AdminServiceCreateProjectWhitelistedDomainRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_alert

> serde_json::Value admin_service_delete_alert(organization, project, name)
DeleteAlert deletes the virtual file for a UI-managed alert, triggers a reconcile, and waits for the alert to be deleted in the runtime

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_organization

> serde_json::Value admin_service_delete_organization(name)
DeleteOrganization deletes an organizations

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**name** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_project

> models::V1DeleteProjectResponse admin_service_delete_project(organization_name, name)
DeleteProject deletes an project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | **String** |  | [required] |

### Return type

[**models::V1DeleteProjectResponse**](v1DeleteProjectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_report

> serde_json::Value admin_service_delete_report(organization, project, name)
DeleteReport deletes the virtual file for a UI-managed report, triggers a reconcile, and waits for the report to be deleted in the runtime

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_service

> models::V1DeleteServiceResponse admin_service_delete_service(organization_name, name)
DeleteService deletes a service per organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | **String** |  | [required] |

### Return type

[**models::V1DeleteServiceResponse**](v1DeleteServiceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_delete_usergroup

> serde_json::Value admin_service_delete_usergroup(organization, usergroup)
DeleteUsergroup deletes the user group from the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_deny_project_access

> serde_json::Value admin_service_deny_project_access(id, body)


### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**id** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_edit_alert

> serde_json::Value admin_service_edit_alert(organization, project, name, body)
EditAlert edits a virtual file for a UI-managed alert, triggers a reconcile, and waits for the alert to be updated in the runtime

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | [**AdminServiceCreateAlertRequest**](AdminServiceCreateAlertRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_edit_report

> serde_json::Value admin_service_edit_report(organization, project, name, body)
EditReport edits a virtual file for a UI-managed report, triggers a reconcile, and waits for the report to be updated in the runtime

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | [**AdminServiceCreateReportRequest**](AdminServiceCreateReportRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_edit_usergroup

> serde_json::Value admin_service_edit_usergroup(organization, usergroup, body)
EditUsergroup renames the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceEditUsergroupRequest**](AdminServiceEditUsergroupRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_generate_alert_yaml

> models::V1GenerateAlertYamlResponse admin_service_generate_alert_yaml(organization, project, body)
GenerateAlertYAML generates YAML for an alert to be copied into a project's Git repository

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceCreateAlertRequest**](AdminServiceCreateAlertRequest.md) |  | [required] |

### Return type

[**models::V1GenerateAlertYamlResponse**](v1GenerateAlertYAMLResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_generate_report_yaml

> models::V1GenerateReportYamlResponse admin_service_generate_report_yaml(organization, project, body)
GenerateReportYAML generates YAML for a scheduled report to be copied into a project's Git repository

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceCreateReportRequest**](AdminServiceCreateReportRequest.md) |  | [required] |

### Return type

[**models::V1GenerateReportYamlResponse**](v1GenerateReportYAMLResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_alert_meta

> models::V1GetAlertMetaResponse admin_service_get_alert_meta(project_id, body)
GetAlertMeta returns metadata for checking an alert. It's currently only called by the alert reconciler in the runtime.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**body** | [**AdminServiceGetAlertMetaRequest**](AdminServiceGetAlertMetaRequest.md) |  | [required] |

### Return type

[**models::V1GetAlertMetaResponse**](v1GetAlertMetaResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_alert_yaml

> models::V1GetAlertYamlResponse admin_service_get_alert_yaml(organization, project, name)
GenerateAlertYAML generates YAML for an alert to be copied into a project's Git repository

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |

### Return type

[**models::V1GetAlertYamlResponse**](v1GetAlertYAMLResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_billing_project_credentials

> models::V1GetBillingProjectCredentialsResponse admin_service_get_billing_project_credentials(body)
GetBillingProjectCredentials returns credentials for the configured cloud metrics project filtered by request organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1GetBillingProjectCredentialsRequest**](V1GetBillingProjectCredentialsRequest.md) |  | [required] |

### Return type

[**models::V1GetBillingProjectCredentialsResponse**](v1GetBillingProjectCredentialsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_billing_subscription

> models::V1GetBillingSubscriptionResponse admin_service_get_billing_subscription(organization)
GetBillingSubscription lists the subscription for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |

### Return type

[**models::V1GetBillingSubscriptionResponse**](v1GetBillingSubscriptionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_bookmark

> models::V1GetBookmarkResponse admin_service_get_bookmark(bookmark_id)
GetBookmark returns the bookmark for the given user for the given project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**bookmark_id** | **String** |  | [required] |

### Return type

[**models::V1GetBookmarkResponse**](v1GetBookmarkResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_clone_credentials

> models::V1GetCloneCredentialsResponse admin_service_get_clone_credentials(organization, project)
GetCloneCredentials returns credentials and other details for a project's Git repository or archive path if git repo is not configured.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |

### Return type

[**models::V1GetCloneCredentialsResponse**](v1GetCloneCredentialsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_current_magic_auth_token

> models::V1GetCurrentMagicAuthTokenResponse admin_service_get_current_magic_auth_token()
GetCurrentMagicAuthToken returns information about the current magic auth token.

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1GetCurrentMagicAuthTokenResponse**](v1GetCurrentMagicAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_current_user

> models::V1GetCurrentUserResponse admin_service_get_current_user()
GetCurrentUser returns the currently authenticated user (if any)

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1GetCurrentUserResponse**](v1GetCurrentUserResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_deployment_credentials

> models::V1GetDeploymentCredentialsResponse admin_service_get_deployment_credentials(organization, project, body)
GetDeploymentCredentials returns runtime info and access token on behalf of a specific user, or alternatively for a raw set of JWT attributes

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | [**AdminServiceGetDeploymentCredentialsRequest**](AdminServiceGetDeploymentCredentialsRequest.md) |  | [required] |

### Return type

[**models::V1GetDeploymentCredentialsResponse**](v1GetDeploymentCredentialsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_github_repo_status

> models::V1GetGithubRepoStatusResponse admin_service_get_github_repo_status(github_url)
GetGithubRepoRequest returns info about a Github repo based on the caller's installations. If the caller has not granted access to the repository, instructions for granting access are returned.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**github_url** | Option<**String**> |  |  |

### Return type

[**models::V1GetGithubRepoStatusResponse**](v1GetGithubRepoStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_github_user_status

> models::V1GetGithubUserStatusResponse admin_service_get_github_user_status()
GetGithubUserStatus returns info about a Github user account based on the caller's installations. If we don't have access to user's personal account tokens or it is expired, instructions for granting access are returned.

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1GetGithubUserStatusResponse**](v1GetGithubUserStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_i_frame

> models::V1GetIFrameResponse admin_service_get_i_frame(organization, project, body)
GetIFrame returns the iframe URL for the given project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** | Organization that owns the project to embed. | [required] |
**project** | **String** | Project that has the resource(s) to embed. | [required] |
**body** | [**AdminServiceGetIFrameRequest**](AdminServiceGetIFrameRequest.md) |  | [required] |

### Return type

[**models::V1GetIFrameResponse**](v1GetIFrameResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_organization

> models::V1GetOrganizationResponse admin_service_get_organization(name)
GetOrganization returns information about a specific organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**name** | **String** |  | [required] |

### Return type

[**models::V1GetOrganizationResponse**](v1GetOrganizationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_organization_name_for_domain

> models::V1GetOrganizationNameForDomainResponse admin_service_get_organization_name_for_domain(domain)
GetOrganizationNameForDomain finds the org name for a custom domain. If the application detects it is running on a non-default domain, it can use this to find the org to present. It can be called without being authenticated.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**domain** | **String** |  | [required] |

### Return type

[**models::V1GetOrganizationNameForDomainResponse**](v1GetOrganizationNameForDomainResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_payments_portal_url

> models::V1GetPaymentsPortalUrlResponse admin_service_get_payments_portal_url(organization, return_url)
GetPaymentsPortalURL returns the URL for the billing session to collect payment method

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**return_url** | Option<**String**> |  |  |

### Return type

[**models::V1GetPaymentsPortalUrlResponse**](v1GetPaymentsPortalURLResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_project

> models::V1GetProjectResponse admin_service_get_project(organization_name, name, access_token_ttl_seconds, issue_superuser_token)
GetProject returns information about a specific project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | **String** |  | [required] |
**access_token_ttl_seconds** | Option<**i64**> |  |  |
**issue_superuser_token** | Option<**bool**> |  |  |

### Return type

[**models::V1GetProjectResponse**](v1GetProjectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_project_access_request

> models::V1GetProjectAccessRequestResponse admin_service_get_project_access_request(id)


### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**id** | **String** |  | [required] |

### Return type

[**models::V1GetProjectAccessRequestResponse**](v1GetProjectAccessRequestResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_project_by_id

> models::V1GetProjectByIdResponse admin_service_get_project_by_id(id)
GetProject returns information about a specific project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**id** | **String** |  | [required] |

### Return type

[**models::V1GetProjectByIdResponse**](v1GetProjectByIDResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_project_variables

> models::V1GetProjectVariablesResponse admin_service_get_project_variables(organization, project, environment, for_all_environments)
GetProjectVariables returns project variables.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** | Organization the project belongs to. | [required] |
**project** | **String** | Project to get variables for. | [required] |
**environment** | Option<**String**> | Environment to get the variables for. If empty, only variables shared across all environments are returned. |  |
**for_all_environments** | Option<**bool**> | If true, return variable values for all environments. Can't be used together with the \"environment\" option. |  |

### Return type

[**models::V1GetProjectVariablesResponse**](v1GetProjectVariablesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_repo_meta

> models::V1GetRepoMetaResponse admin_service_get_repo_meta(project_id, branch)
GetRepoMeta returns credentials and other metadata for accessing a project's repo

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**branch** | Option<**String**> |  |  |

### Return type

[**models::V1GetRepoMetaResponse**](v1GetRepoMetaResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_report_meta

> models::V1GetReportMetaResponse admin_service_get_report_meta(project_id, body)
GetReportMeta returns metadata for generating a report. It's currently only called by the report reconciler in the runtime.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**body** | [**AdminServiceGetReportMetaRequest**](AdminServiceGetReportMetaRequest.md) |  | [required] |

### Return type

[**models::V1GetReportMetaResponse**](v1GetReportMetaResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_user

> models::V1GetUserResponse admin_service_get_user(email)
GetUser returns user by email

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**email** | Option<**String**> |  |  |

### Return type

[**models::V1GetUserResponse**](v1GetUserResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_get_usergroup

> models::V1GetUsergroupResponse admin_service_get_usergroup(organization, usergroup, page_size, page_token)
GetUsergroups returns the user group details

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1GetUsergroupResponse**](v1GetUsergroupResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_hibernate_project

> serde_json::Value admin_service_hibernate_project(organization, project)
HibernateProject hibernates a project by tearing down its deployments.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_issue_magic_auth_token

> models::V1IssueMagicAuthTokenResponse admin_service_issue_magic_auth_token(organization, project, body)
IssueMagicAuthToken creates a \"magic\" auth token that provides limited access to a specific filtered dashboard in a specific project.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** | Organization that owns the project. | [required] |
**project** | **String** | Project to create the magic auth token in. | [required] |
**body** | [**AdminServiceIssueMagicAuthTokenRequest**](AdminServiceIssueMagicAuthTokenRequest.md) |  | [required] |

### Return type

[**models::V1IssueMagicAuthTokenResponse**](v1IssueMagicAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_issue_representative_auth_token

> models::V1IssueRepresentativeAuthTokenResponse admin_service_issue_representative_auth_token(body)
IssueRepresentativeAuthToken returns the temporary token for given email

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1IssueRepresentativeAuthTokenRequest**](V1IssueRepresentativeAuthTokenRequest.md) |  | [required] |

### Return type

[**models::V1IssueRepresentativeAuthTokenResponse**](v1IssueRepresentativeAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_issue_service_auth_token

> models::V1IssueServiceAuthTokenResponse admin_service_issue_service_auth_token(organization_name, service_name, body)
IssueServiceAuthToken returns the temporary token for given service account

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**service_name** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**models::V1IssueServiceAuthTokenResponse**](v1IssueServiceAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_leave_organization

> serde_json::Value admin_service_leave_organization(organization)
LeaveOrganization removes the current user from the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_bookmarks

> models::V1ListBookmarksResponse admin_service_list_bookmarks(project_id, resource_kind, resource_name)
ListBookmarks lists all the bookmarks for the user and global ones for dashboard

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | Option<**String**> |  |  |
**resource_kind** | Option<**String**> |  |  |
**resource_name** | Option<**String**> |  |  |

### Return type

[**models::V1ListBookmarksResponse**](v1ListBookmarksResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_github_user_repos

> models::V1ListGithubUserReposResponse admin_service_list_github_user_repos()


### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1ListGithubUserReposResponse**](v1ListGithubUserReposResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_magic_auth_tokens

> models::V1ListMagicAuthTokensResponse admin_service_list_magic_auth_tokens(organization, project, page_size, page_token)
ListMagicAuthTokens lists all the magic auth tokens for a specific project.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListMagicAuthTokensResponse**](v1ListMagicAuthTokensResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_organization_billing_issues

> models::V1ListOrganizationBillingIssuesResponse admin_service_list_organization_billing_issues(organization)
ListOrganizationBillingIssues lists all the billing issues for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |

### Return type

[**models::V1ListOrganizationBillingIssuesResponse**](v1ListOrganizationBillingIssuesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_organization_invites

> models::V1ListOrganizationInvitesResponse admin_service_list_organization_invites(organization, page_size, page_token)
ListOrganizationInvites lists all the org invites

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListOrganizationInvitesResponse**](v1ListOrganizationInvitesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_organization_member_usergroups

> models::V1ListOrganizationMemberUsergroupsResponse admin_service_list_organization_member_usergroups(organization, page_size, page_token)
ListOrganizationMemberUsergroups lists the org's user groups

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListOrganizationMemberUsergroupsResponse**](v1ListOrganizationMemberUsergroupsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_organization_member_users

> models::V1ListOrganizationMemberUsersResponse admin_service_list_organization_member_users(organization, page_size, page_token)
ListOrganizationMemberUsers lists all the org members

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListOrganizationMemberUsersResponse**](v1ListOrganizationMemberUsersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_organizations

> models::V1ListOrganizationsResponse admin_service_list_organizations(page_size, page_token)
ListOrganizations lists all the organizations currently managed by the admin

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListOrganizationsResponse**](v1ListOrganizationsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_project_invites

> models::V1ListProjectInvitesResponse admin_service_list_project_invites(organization, project, page_size, page_token)
ListProjectInvites lists all the project invites

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListProjectInvitesResponse**](v1ListProjectInvitesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_project_member_usergroups

> models::V1ListProjectMemberUsergroupsResponse admin_service_list_project_member_usergroups(organization, project, page_size, page_token)
ListProjectMemberUsergroups lists the org's user groups

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListProjectMemberUsergroupsResponse**](v1ListProjectMemberUsergroupsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_project_member_users

> models::V1ListProjectMemberUsersResponse admin_service_list_project_member_users(organization, project, page_size, page_token)
ListProjectMemberUsers lists all the project members

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListProjectMemberUsersResponse**](v1ListProjectMemberUsersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_project_whitelisted_domains

> models::V1ListProjectWhitelistedDomainsResponse admin_service_list_project_whitelisted_domains(organization, project)
ListWhitelistedDomains lists all the whitelisted domains of the project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |

### Return type

[**models::V1ListProjectWhitelistedDomainsResponse**](v1ListProjectWhitelistedDomainsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_projects_for_organization

> models::V1ListProjectsForOrganizationResponse admin_service_list_projects_for_organization(organization_name, page_size, page_token)
ListProjectsForOrganization lists all the projects currently available for given organizations

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListProjectsForOrganizationResponse**](v1ListProjectsForOrganizationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_public_billing_plans

> models::V1ListPublicBillingPlansResponse admin_service_list_public_billing_plans()
ListPublicBillingPlans lists all public billing plans

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1ListPublicBillingPlansResponse**](v1ListPublicBillingPlansResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_service_auth_tokens

> models::V1ListServiceAuthTokensResponse admin_service_list_service_auth_tokens(organization_name, service_name)
ListServiceAuthTokens lists all the service auth tokens

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**service_name** | **String** |  | [required] |

### Return type

[**models::V1ListServiceAuthTokensResponse**](v1ListServiceAuthTokensResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_services

> models::V1ListServicesResponse admin_service_list_services(organization_name)
ListService returns all the services per organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |

### Return type

[**models::V1ListServicesResponse**](v1ListServicesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_superusers

> models::V1ListSuperusersResponse admin_service_list_superusers()
ListSuperusers lists all the superusers

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1ListSuperusersResponse**](v1ListSuperusersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_usergroup_member_users

> models::V1ListUsergroupMemberUsersResponse admin_service_list_usergroup_member_users(organization, usergroup, page_size, page_token)
ListUsergroupMemberUsers lists all the user group members

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListUsergroupMemberUsersResponse**](v1ListUsergroupMemberUsersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_list_whitelisted_domains

> models::V1ListWhitelistedDomainsResponse admin_service_list_whitelisted_domains(organization)
ListWhitelistedDomains lists all the whitelisted domains for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |

### Return type

[**models::V1ListWhitelistedDomainsResponse**](v1ListWhitelistedDomainsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_ping

> models::V1PingResponse admin_service_ping()
Ping returns information about the server

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1PingResponse**](v1PingResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_provision

> models::V1ProvisionResponse admin_service_provision(deployment_id, body)
Provision provisions a new resource for a deployment. If an existing resource matches the request, it will be returned without provisioning a new resource.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**deployment_id** | **String** | Deployment to provision a resource for. If it's blank and the request is made with a deployment access token, the deployment is inferred from the token. | [required] |
**body** | [**AdminServiceProvisionRequest**](AdminServiceProvisionRequest.md) |  | [required] |

### Return type

[**models::V1ProvisionResponse**](v1ProvisionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_pull_virtual_repo

> models::V1PullVirtualRepoResponse admin_service_pull_virtual_repo(project_id, branch, page_size, page_token)
PullVirtualRepo fetches files from a project's virtual repo

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**project_id** | **String** |  | [required] |
**branch** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1PullVirtualRepoResponse**](v1PullVirtualRepoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_redeploy_project

> serde_json::Value admin_service_redeploy_project(organization, project)
RedeployProject creates a new production deployment for a project. If the project currently has another production deployment, the old deployment will be deprovisioned. This RPC can be used to redeploy a project that has been hibernated.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_bookmark

> serde_json::Value admin_service_remove_bookmark(bookmark_id)
RemoveBookmark removes the bookmark for the given user or all users

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**bookmark_id** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_organization_member_user

> serde_json::Value admin_service_remove_organization_member_user(organization, email, keep_project_roles)
RemoveOrganizationMemberUser removes member from the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**email** | **String** |  | [required] |
**keep_project_roles** | Option<**bool**> |  |  |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_organization_member_usergroup

> serde_json::Value admin_service_remove_organization_member_usergroup(organization, usergroup)
RemoveOrganizationMemberUsergroup revokes the organization-level role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_project_member_user

> serde_json::Value admin_service_remove_project_member_user(organization, project, email)
RemoveProjectMemberUser removes member from the project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**email** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_project_member_usergroup

> serde_json::Value admin_service_remove_project_member_usergroup(organization, project, usergroup)
RemoveProjectMemberUsergroup revokes the project-level role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_project_whitelisted_domain

> serde_json::Value admin_service_remove_project_whitelisted_domain(organization, project, domain)
RemoveProjectWhitelistedDomain removes a domain from the project's whitelisted

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**domain** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_usergroup_member_user

> serde_json::Value admin_service_remove_usergroup_member_user(organization, usergroup, email)
RemoveUsergroupMemberUser removes member from the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**email** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_remove_whitelisted_domain

> serde_json::Value admin_service_remove_whitelisted_domain(organization, domain)
RemoveWhitelistedDomain removes a domain from the whitelist list

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**domain** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_rename_usergroup

> serde_json::Value admin_service_rename_usergroup(organization, usergroup, body)
RenameUsergroup renames the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceCreateUsergroupRequest**](AdminServiceCreateUsergroupRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_renew_billing_subscription

> models::V1RenewBillingSubscriptionResponse admin_service_renew_billing_subscription(organization, body)
RenewBillingSubscription renews the billing plan for the organization once cancelled

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**body** | [**AdminServiceUpdateBillingSubscriptionRequest**](AdminServiceUpdateBillingSubscriptionRequest.md) |  | [required] |

### Return type

[**models::V1RenewBillingSubscriptionResponse**](v1RenewBillingSubscriptionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_request_project_access

> serde_json::Value admin_service_request_project_access(organization, project, body)


### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_revoke_current_auth_token

> models::V1RevokeCurrentAuthTokenResponse admin_service_revoke_current_auth_token()
RevokeCurrentAuthToken revoke the current auth token

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1RevokeCurrentAuthTokenResponse**](v1RevokeCurrentAuthTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_revoke_magic_auth_token

> serde_json::Value admin_service_revoke_magic_auth_token(token_id)
RevokeMagicAuthToken revokes a magic auth token.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**token_id** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_revoke_service_auth_token

> serde_json::Value admin_service_revoke_service_auth_token(token_id)
RevokeServiceAuthToken revoke the service auth token

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**token_id** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_search_project_names

> models::V1SearchProjectNamesResponse admin_service_search_project_names(name_pattern, annotations, page_size, page_token)
SearchProjectNames returns project names matching the pattern

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**name_pattern** | Option<**String**> |  |  |
**annotations** | Option<**String**> | This is a request variable of the map type. The query format is \"map_name[key]=value\", e.g. If the map name is Age, the key type is string, and the value type is integer, the query parameter is expressed as Age[\"bob\"]=18 |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1SearchProjectNamesResponse**](v1SearchProjectNamesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_search_project_users

> models::V1SearchProjectUsersResponse admin_service_search_project_users(organization, project, email_query, page_size, page_token)
SearchProjectUsers returns users who has access to to a project (including org members that have access through a usergroup)

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**email_query** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1SearchProjectUsersResponse**](v1SearchProjectUsersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_search_users

> models::V1SearchUsersResponse admin_service_search_users(email_pattern, page_size, page_token)
GetUsersByEmail returns users by email

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**email_pattern** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1SearchUsersResponse**](v1SearchUsersResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_set_organization_member_user_role

> serde_json::Value admin_service_set_organization_member_user_role(organization, email, body)
SetOrganizationMemberUserRole sets the role for the member

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**email** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_set_organization_member_usergroup_role

> serde_json::Value admin_service_set_organization_member_usergroup_role(organization, usergroup, body)
SetOrganizationMemberUsergroupRole sets the role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_set_project_member_user_role

> serde_json::Value admin_service_set_project_member_user_role(organization, project, email, body)
SetProjectMemberUserRole sets the role for the member

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**email** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_set_project_member_usergroup_role

> serde_json::Value admin_service_set_project_member_usergroup_role(organization, project, usergroup, body)
SetProjectMemberUsergroupRole sets the role for the user group

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**usergroup** | **String** |  | [required] |
**body** | [**AdminServiceSetOrganizationMemberUserRoleRequest**](AdminServiceSetOrganizationMemberUserRoleRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_set_superuser

> serde_json::Value admin_service_set_superuser(body)
SetSuperuser adds/remove a superuser

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SetSuperuserRequest**](V1SetSuperuserRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_delete_organization_billing_issue

> serde_json::Value admin_service_sudo_delete_organization_billing_issue(organization, r#type)
SudoDeleteOrganizationBillingIssue deletes a billing issue of a type for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**r#type** | **String** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_extend_trial

> models::V1SudoExtendTrialResponse admin_service_sudo_extend_trial(body)
SudoExtendTrial extends the trial period for an organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoExtendTrialRequest**](V1SudoExtendTrialRequest.md) |  | [required] |

### Return type

[**models::V1SudoExtendTrialResponse**](v1SudoExtendTrialResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_get_resource

> models::V1SudoGetResourceResponse admin_service_sudo_get_resource(user_id, org_id, project_id, deployment_id, instance_id)
SudoGetResource returns details about a resource by ID lookup

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**user_id** | Option<**String**> |  |  |
**org_id** | Option<**String**> |  |  |
**project_id** | Option<**String**> |  |  |
**deployment_id** | Option<**String**> |  |  |
**instance_id** | Option<**String**> |  |  |

### Return type

[**models::V1SudoGetResourceResponse**](v1SudoGetResourceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_issue_runtime_manager_token

> models::V1SudoIssueRuntimeManagerTokenResponse admin_service_sudo_issue_runtime_manager_token(body)
SudoIssueRuntimeManagerToken returns a runtime JWT with full manager permissions for a runtime.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoIssueRuntimeManagerTokenRequest**](V1SudoIssueRuntimeManagerTokenRequest.md) |  | [required] |

### Return type

[**models::V1SudoIssueRuntimeManagerTokenResponse**](v1SudoIssueRuntimeManagerTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_trigger_billing_repair

> serde_json::Value admin_service_sudo_trigger_billing_repair(body)
SudoTriggerBillingRepair triggers billing repair jobs for orgs that doesn't have billing info and puts them on trial

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_update_annotations

> models::V1SudoUpdateAnnotationsResponse admin_service_sudo_update_annotations(body)
SudoUpdateAnnotations endpoint for superusers to update project annotations

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoUpdateAnnotationsRequest**](V1SudoUpdateAnnotationsRequest.md) |  | [required] |

### Return type

[**models::V1SudoUpdateAnnotationsResponse**](v1SudoUpdateAnnotationsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_update_organization_billing_customer

> models::V1SudoUpdateOrganizationBillingCustomerResponse admin_service_sudo_update_organization_billing_customer(body)
SudoUpdateOrganizationBillingCustomer update the billing customer for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoUpdateOrganizationBillingCustomerRequest**](V1SudoUpdateOrganizationBillingCustomerRequest.md) |  | [required] |

### Return type

[**models::V1SudoUpdateOrganizationBillingCustomerResponse**](v1SudoUpdateOrganizationBillingCustomerResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_update_organization_custom_domain

> models::V1SudoUpdateOrganizationCustomDomainResponse admin_service_sudo_update_organization_custom_domain(body)
SudoUpdateOrganizationCustomDomain updates the custom domain for an organization. It only updates the custom domain in the database, which is used to ensure correct redirects. The DNS records and ingress TLS must be configured separately.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoUpdateOrganizationCustomDomainRequest**](V1SudoUpdateOrganizationCustomDomainRequest.md) |  | [required] |

### Return type

[**models::V1SudoUpdateOrganizationCustomDomainResponse**](v1SudoUpdateOrganizationCustomDomainResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_update_organization_quotas

> models::V1SudoUpdateOrganizationQuotasResponse admin_service_sudo_update_organization_quotas(body)
SudoUpdateOrganizationQuotas update the quotas available for orgs

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoUpdateOrganizationQuotasRequest**](V1SudoUpdateOrganizationQuotasRequest.md) |  | [required] |

### Return type

[**models::V1SudoUpdateOrganizationQuotasResponse**](v1SudoUpdateOrganizationQuotasResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_sudo_update_user_quotas

> models::V1SudoUpdateUserQuotasResponse admin_service_sudo_update_user_quotas(body)
SudoUpdateUserQuotas update the quotas for users

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1SudoUpdateUserQuotasRequest**](V1SudoUpdateUserQuotasRequest.md) |  | [required] |

### Return type

[**models::V1SudoUpdateUserQuotasResponse**](v1SudoUpdateUserQuotasResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_trigger_reconcile

> serde_json::Value admin_service_trigger_reconcile(deployment_id, body)
TriggerReconcile triggers reconcile for the project's prod deployment. DEPRECATED: Clients should call CreateTrigger directly on the deployed runtime instead.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**deployment_id** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_trigger_redeploy

> serde_json::Value admin_service_trigger_redeploy(body)
TriggerRedeploy is similar to RedeployProject. DEPRECATED: Use RedeployProject instead.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1TriggerRedeployRequest**](V1TriggerRedeployRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_trigger_refresh_sources

> serde_json::Value admin_service_trigger_refresh_sources(deployment_id, body)
TriggerRefreshSources refresh the source for production deployment. DEPRECATED: Clients should call CreateTrigger directly on the deployed runtime instead.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**deployment_id** | **String** |  | [required] |
**body** | [**AdminServiceTriggerRefreshSourcesRequest**](AdminServiceTriggerRefreshSourcesRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_trigger_report

> serde_json::Value admin_service_trigger_report(organization, project, name, body)
TriggerReport triggers an ad-hoc report run

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_unsubscribe_alert

> serde_json::Value admin_service_unsubscribe_alert(organization, project, name, body)
UnsubscribeAlert removes the calling user from a alert's recipients list

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_unsubscribe_report

> serde_json::Value admin_service_unsubscribe_report(organization, project, name, body)
UnsubscribeReport removes the calling user from a reports recipients list

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_billing_subscription

> models::V1UpdateBillingSubscriptionResponse admin_service_update_billing_subscription(organization, body)
UpdateBillingSubscription updates the billing plan for the organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**body** | [**AdminServiceUpdateBillingSubscriptionRequest**](AdminServiceUpdateBillingSubscriptionRequest.md) |  | [required] |

### Return type

[**models::V1UpdateBillingSubscriptionResponse**](v1UpdateBillingSubscriptionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_bookmark

> serde_json::Value admin_service_update_bookmark(body)
UpdateBookmark updates a bookmark for the given user for the given project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1UpdateBookmarkRequest**](V1UpdateBookmarkRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_organization

> models::V1UpdateOrganizationResponse admin_service_update_organization(name, body)
UpdateOrganization deletes an organizations

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**name** | **String** |  | [required] |
**body** | [**AdminServiceUpdateOrganizationRequest**](AdminServiceUpdateOrganizationRequest.md) |  | [required] |

### Return type

[**models::V1UpdateOrganizationResponse**](v1UpdateOrganizationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_project

> models::V1UpdateProjectResponse admin_service_update_project(organization_name, name, body)
UpdateProject updates a project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | [**AdminServiceUpdateProjectRequest**](AdminServiceUpdateProjectRequest.md) |  | [required] |

### Return type

[**models::V1UpdateProjectResponse**](v1UpdateProjectResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_project_variables

> models::V1UpdateProjectVariablesResponse admin_service_update_project_variables(organization, project, body)
UpdateProjectVariables updates variables for a project.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** | Organization the project belongs to. | [required] |
**project** | **String** | Project to update variables for. | [required] |
**body** | [**AdminServiceUpdateProjectVariablesRequest**](AdminServiceUpdateProjectVariablesRequest.md) |  | [required] |

### Return type

[**models::V1UpdateProjectVariablesResponse**](v1UpdateProjectVariablesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_service

> models::V1UpdateServiceResponse admin_service_update_service(organization_name, name, body)
UpdateService updates a service per organization

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization_name** | **String** |  | [required] |
**name** | **String** |  | [required] |
**body** | [**AdminServiceUpdateServiceRequest**](AdminServiceUpdateServiceRequest.md) |  | [required] |

### Return type

[**models::V1UpdateServiceResponse**](v1UpdateServiceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_update_user_preferences

> models::V1UpdateUserPreferencesResponse admin_service_update_user_preferences(body)
UpdateUserPreferences updates the preferences for the user

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1UpdateUserPreferencesRequest**](V1UpdateUserPreferencesRequest.md) |  | [required] |

### Return type

[**models::V1UpdateUserPreferencesResponse**](v1UpdateUserPreferencesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## admin_service_upload_project_assets

> serde_json::Value admin_service_upload_project_assets(organization, project, body)
Converts a project connected to github to a rill managed project. Uploads the current project to assets.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**organization** | **String** |  | [required] |
**project** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

