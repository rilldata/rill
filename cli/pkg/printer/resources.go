package printer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lensesio/tableprinter"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
)

func (p *Printer) PrintOrgs(orgs []*adminv1.Organization, defaultOrg string) {
	if len(orgs) == 0 {
		p.PrintfWarn("No organizations found\n")
		return
	}

	p.PrintData(toOrgsTable(orgs, defaultOrg))
}

func toOrgsTable(orgs []*adminv1.Organization, defaultOrg string) []*organization {
	res := make([]*organization, 0, len(orgs))

	for _, org := range orgs {
		if strings.EqualFold(org.Name, defaultOrg) {
			org.Name += " (default)"
		}
		res = append(res, toOrgRow(org))
	}

	return res
}

func toOrgRow(o *adminv1.Organization) *organization {
	return &organization{
		Name:      o.Name,
		CreatedAt: o.CreatedOn.AsTime().Local().Format(time.DateTime),
	}
}

type organization struct {
	Name      string `header:"name" json:"name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}

func (p *Printer) PrintProjects(projs []*adminv1.Project) {
	if len(projs) == 0 {
		p.PrintfWarn("No projects found\n")
		return
	}

	p.PrintData(toProjectTable(projs))
}

func toProjectTable(projects []*adminv1.Project) []*project {
	projs := make([]*project, 0, len(projects))

	for _, proj := range projects {
		projs = append(projs, toProjectRow(proj))
	}

	return projs
}

func toProjectRow(o *adminv1.Project) *project {
	var githubURL string
	if o.ManagedGitId == "" {
		githubURL = strings.TrimSuffix(o.GitRemote, ".git")
		if o.Subpath != "" {
			githubURL = filepath.Join(githubURL, "tree", o.ProdBranch, o.Subpath)
		}
	}

	return &project{
		Name:         o.Name,
		Public:       o.Public,
		GithubURL:    githubURL,
		Organization: o.OrgName,
	}
}

type project struct {
	Name         string `header:"name" json:"name"`
	Public       bool   `header:"public" json:"public"`
	GithubURL    string `header:"github" json:"github"`
	Organization string `header:"organization" json:"organization"`
}

func (p *Printer) PrintUsers(users []*adminv1.User) {
	if len(users) == 0 {
		p.PrintfWarn("No users found\n")
		return
	}

	p.PrintData(toUsersTable(users))
}

func toUsersTable(users []*adminv1.User) []*user {
	allUsers := make([]*user, 0, len(users))

	for _, m := range users {
		allUsers = append(allUsers, toUserRow(m))
	}

	return allUsers
}

func toUserRow(m *adminv1.User) *user {
	return &user{
		Name:  m.DisplayName,
		Email: m.Email,
	}
}

type user struct {
	Name  string `header:"name" json:"display_name"`
	Email string `header:"email" json:"email"`
}

func (p *Printer) PrintOrganizationMemberUsers(members []*adminv1.OrganizationMemberUser) {
	if len(members) == 0 {
		p.PrintfWarn("No members found\n")
		return
	}

	allMembers := make([]*memberUserWithRole, 0, len(members))
	for _, m := range members {
		memberAttrs := ""
		if m.Attributes != nil && len(m.Attributes.Fields) > 0 {
			attrMap := m.Attributes.AsMap()
			var attrs []string
			for key, value := range attrMap {
				attrs = append(attrs, fmt.Sprintf("%s=%v", key, value))
			}
			memberAttrs = strings.Join(attrs, ", ")
		}

		allMembers = append(allMembers, &memberUserWithRole{
			Email:      m.UserEmail,
			Name:       m.UserName,
			RoleName:   m.RoleName,
			Attributes: memberAttrs,
		})
	}

	p.PrintData(allMembers)
}

func (p *Printer) PrintProjectMemberUsers(members []*adminv1.ProjectMemberUser) {
	if len(members) == 0 {
		p.PrintfWarn("No members found\n")
		return
	}

	allMembers := make([]*projectMemberUserWithRole, 0, len(members))
	for _, m := range members {
		allMembers = append(allMembers, &projectMemberUserWithRole{
			Email:    m.UserEmail,
			Name:     m.UserName,
			RoleName: m.RoleName,
		})
	}

	p.PrintData(allMembers)
}

func (p *Printer) PrintOrganizationMemberServices(members []*adminv1.OrganizationMemberService) {
	if len(members) == 0 {
		p.PrintfWarn("No services found\n")
		return
	}

	allMembers := make([]*orgMemberService, 0, len(members))
	for _, m := range members {
		attrBytes, err := json.Marshal(m.Attributes)
		if err != nil {
			panic(fmt.Errorf("failed to marshal service attributes: %w", err))
		}
		allMembers = append(allMembers, &orgMemberService{
			Name:            m.Name,
			RoleName:        m.RoleName,
			HasProjectRoles: m.HasProjectRoles,
			Attributes:      string(attrBytes),
		})
	}

	p.PrintData(allMembers)
}

func (p *Printer) PrintProjectMemberServices(members []*adminv1.ProjectMemberService) {
	if len(members) == 0 {
		p.PrintfWarn("No services found\n")
		return
	}

	allMembers := make([]*projectMemberService, 0, len(members))
	for _, m := range members {
		attrBytes, err := json.Marshal(m.Attributes)
		if err != nil {
			panic(fmt.Errorf("failed to marshal service attributes: %w", err))
		}

		allMembers = append(allMembers, &projectMemberService{
			Name:            m.Name,
			ProjectName:     m.ProjectName,
			ProjectRoleName: m.ProjectRoleName,
			OrgRoleName:     m.OrgRoleName,
			Attributes:      string(attrBytes),
		})
	}

	p.PrintData(allMembers)
}

func (p *Printer) PrintUsergroupMemberUsers(members []*adminv1.UsergroupMemberUser) {
	if len(members) == 0 {
		p.PrintfWarn("No members found\n")
		return
	}

	allMembers := make([]*memberUser, 0, len(members))
	for _, m := range members {
		allMembers = append(allMembers, &memberUser{
			Email: m.UserEmail,
			Name:  m.UserName,
		})
	}

	p.PrintData(allMembers)
}

type memberUser struct {
	Email string `header:"email" json:"email"`
	Name  string `header:"name" json:"display_name"`
}

type memberUserWithRole struct {
	Email      string `header:"email" json:"email"`
	Name       string `header:"name" json:"display_name"`
	RoleName   string `header:"role" json:"role_name"`
	Attributes string `header:"attributes" json:"attributes"`
}

type projectMemberUserWithRole struct {
	Email    string `header:"email" json:"email"`
	Name     string `header:"name" json:"display_name"`
	RoleName string `header:"role" json:"role_name"`
}

type orgMemberService struct {
	Name            string `header:"name" json:"name"`
	RoleName        string `header:"org_role" json:"role_name"`
	HasProjectRoles bool   `header:"has_project_roles" json:"has_project_roles"`
	Attributes      string `header:"attributes" json:"attributes"`
}

type projectMemberService struct {
	Name            string `header:"name" json:"name"`
	ProjectName     string `header:"project" json:"project_name"`
	ProjectRoleName string `header:"project_role" json:"project_role_name"`
	OrgRoleName     string `header:"org_role" json:"org_role_name"`
	Attributes      string `header:"attributes" json:"attributes"`
}

func (p *Printer) PrintOrganizationInvites(invites []*adminv1.OrganizationInvite) {
	if len(invites) == 0 {
		return
	}
	rows := make([]*organizationInvite, 0, len(invites))
	for _, i := range invites {
		rows = append(rows, &organizationInvite{
			Email:     i.Email,
			RoleName:  i.RoleName,
			InvitedBy: i.InvitedBy,
		})
	}
	p.PrintDataWithTitle(rows, "Invites pending acceptance")
}

type organizationInvite struct {
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	InvitedBy string `header:"invited_by" json:"invited_by"`
}

func (p *Printer) PrintProjectInvites(invites []*adminv1.ProjectInvite) {
	if len(invites) == 0 {
		return
	}
	rows := make([]*projectInvite, 0, len(invites))
	for _, i := range invites {
		rows = append(rows, &projectInvite{
			Email:       i.Email,
			RoleName:    i.RoleName,
			OrgRoleName: i.OrgRoleName,
			InvitedBy:   i.InvitedBy,
		})
	}
	p.PrintDataWithTitle(rows, "Invites pending acceptance")
}

type projectInvite struct {
	Email       string `header:"email" json:"email"`
	RoleName    string `header:"role" json:"role_name"`
	OrgRoleName string `header:"org_role" json:"org_role_name"`
	InvitedBy   string `header:"invited_by" json:"invited_by"`
}

func (p *Printer) PrintServices(svcs []*adminv1.Service) {
	if len(svcs) == 0 {
		return
	}
	p.PrintData(toServicesTable(svcs))
}

func toServicesTable(sv []*adminv1.Service) []*service {
	services := make([]*service, 0, len(sv))

	for _, s := range sv {
		services = append(services, toServiceRow(s))
	}

	return services
}

func toServiceRow(s *adminv1.Service) *service {
	attrBytes, err := json.Marshal(s.Attributes)
	if err != nil {
		panic(fmt.Errorf("failed to marshal service attributes: %w", err))
	}

	return &service{
		Name:       s.Name,
		OrgName:    s.OrgName,
		Attributes: string(attrBytes),
		CreatedAt:  s.CreatedOn.AsTime().Local().Format(time.DateTime),
	}
}

type service struct {
	Name       string `header:"name" json:"name"`
	OrgName    string `header:"org_name" json:"org_name"`
	Attributes string `header:"attributes" json:"attributes"`
	CreatedAt  string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}

func (p *Printer) PrintServiceTokens(sts []*adminv1.ServiceToken) {
	if len(sts) == 0 {
		return
	}
	table := make([]*serviceToken, 0, len(sts))
	for _, t := range sts {
		table = append(table, toServiceTokenRow(t))
	}
	p.PrintData(table)
}

func toServiceTokenRow(s *adminv1.ServiceToken) *serviceToken {
	var expiresOn string
	if !s.ExpiresOn.AsTime().IsZero() {
		expiresOn = s.ExpiresOn.AsTime().Local().Format(time.DateTime)
	}

	return &serviceToken{
		ID:        s.Id,
		Prefix:    s.Prefix,
		CreatedOn: s.CreatedOn.AsTime().Local().Format(time.DateTime),
		ExpiresOn: expiresOn,
	}
}

type serviceToken struct {
	ID        string `header:"id" json:"id"`
	Prefix    string `header:"prefix" json:"prefix"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	ExpiresOn string `header:"expires_on,timestamp(ms|utc|human)" json:"expires_on"`
}

func (p *Printer) PrintUserTokens(uts []*adminv1.UserAuthToken) {
	if len(uts) == 0 {
		return
	}
	table := make([]*userToken, 0, len(uts))
	for _, t := range uts {
		table = append(table, toUserTokenRow(t))
	}
	p.PrintData(table)
}

func toUserTokenRow(u *adminv1.UserAuthToken) *userToken {
	var expiresOn, usedOn string
	if u.ExpiresOn != nil {
		expiresOn = u.ExpiresOn.AsTime().Local().Format(time.DateTime)
	}
	if u.UsedOn != nil {
		usedOn = u.UsedOn.AsTime().Local().Format(time.DateTime)
	}

	return &userToken{
		ID:          u.Id,
		ClientName:  u.AuthClientDisplayName,
		Prefix:      u.Prefix,
		Description: u.DisplayName,
		CreatedOn:   u.CreatedOn.AsTime().Local().Format(time.DateTime),
		ExpiresOn:   expiresOn,
		UsedOn:      usedOn,
	}
}

type userToken struct {
	ID          string `header:"id" json:"id"`
	Description string `header:"description" json:"description"`
	Prefix      string `header:"prefix" json:"prefix"`
	ClientName  string `header:"client" json:"client"`
	CreatedOn   string `header:"created,timestamp(ms|utc|human)" json:"created_on"`
	ExpiresOn   string `header:"expires,timestamp(ms|utc|human)" json:"expires_on"`
	UsedOn      string `header:"last used,timestamp(ms|utc|human)" json:"last_used_on"`
}

func (p *Printer) PrintMagicAuthTokens(tkns []*adminv1.MagicAuthToken) {
	if len(tkns) == 0 {
		p.PrintfWarn("No URLs found\n")
		return
	}

	p.PrintData(toMagicAuthTokensTable(tkns))
}

func toMagicAuthTokensTable(tkns []*adminv1.MagicAuthToken) []*magicAuthToken {
	res := make([]*magicAuthToken, 0, len(tkns))

	for _, tkn := range tkns {
		res = append(res, toMagicAuthTokenRow(tkn))
	}

	return res
}

func toMagicAuthTokenRow(t *adminv1.MagicAuthToken) *magicAuthToken {
	expr := metricsview.NewExpressionFromProto(t.Filter)
	filter, err := metricsview.ExpressionToSQL(expr)
	if err != nil {
		panic(err)
	}

	row := &magicAuthToken{
		ID:        t.Id,
		Filter:    filter,
		CreatedBy: t.CreatedByUserEmail,
		CreatedOn: t.CreatedOn.AsTime().Local().Format(time.DateTime),
		UsedOn:    t.UsedOn.AsTime().Local().Format(time.DateTime),
	}
	if t.ExpiresOn != nil {
		row.ExpiresOn = t.ExpiresOn.AsTime().Local().Format(time.DateTime)
	}
	return row
}

type magicAuthToken struct {
	ID        string `header:"id" json:"id"`
	Resource  string `header:"resource" json:"resource"`
	Filter    string `header:"filter" json:"filter"`
	CreatedBy string `header:"created by" json:"created_by"`
	CreatedOn string `header:"created on" json:"created_on"`
	UsedOn    string `header:"last used on" json:"used_on"`
	ExpiresOn string `header:"expires on" json:"expires_on"`
}

func (p *Printer) PrintSubscriptions(subs []*adminv1.Subscription) {
	if len(subs) == 0 {
		return
	}
	p.PrintData(toSubscriptionsTable(subs))
}

func toSubscriptionsTable(subs []*adminv1.Subscription) []*subscription {
	subscriptions := make([]*subscription, 0, len(subs))

	for _, s := range subs {
		subscriptions = append(subscriptions, toSubscriptionRow(s))
	}

	return subscriptions
}

func toSubscriptionRow(s *adminv1.Subscription) *subscription {
	return &subscription{
		ID:                           s.Id,
		PlanName:                     s.Plan.Name,
		PlanDisplayName:              s.Plan.DisplayName,
		StartDate:                    s.StartDate.AsTime().Local().Format(time.DateTime),
		EndDate:                      s.EndDate.AsTime().Local().Format(time.DateTime),
		CurrentBillingCycleStartDate: s.CurrentBillingCycleStartDate.AsTime().Local().Format(time.DateTime),
		CurrentBillingCycleEndDate:   s.CurrentBillingCycleEndDate.AsTime().Local().Format(time.DateTime),
		TrialEndDate:                 s.TrialEndDate.AsTime().Local().Format(time.DateTime),
	}
}

type subscription struct {
	ID                           string `header:"id" json:"id"`
	PlanName                     string `header:"plan_name" json:"plan_name"`
	PlanDisplayName              string `header:"plan_display_name" json:"plan_display_name"`
	StartDate                    string `header:"start_date,timestamp(ms|utc|human)" json:"start_date"`
	EndDate                      string `header:"end_date,timestamp(ms|utc|human)" json:"end_date"`
	CurrentBillingCycleStartDate string `header:"current_billing_cycle_start_date,timestamp(ms|utc|human)" json:"current_billing_cycle_start_date"`
	CurrentBillingCycleEndDate   string `header:"current_billing_cycle_end_date,timestamp(ms|utc|human)" json:"current_billing_cycle_end_date"`
	TrialEndDate                 string `header:"trial_end_date,timestamp(ms|utc|human)" json:"trial_end_date"`
}

func (p *Printer) PrintPlans(plans []*adminv1.BillingPlan) {
	if len(plans) == 0 {
		return
	}
	p.PrintData(toPlansTable(plans))
}

func toPlansTable(plans []*adminv1.BillingPlan) []*plan {
	allPlans := make([]*plan, 0, len(plans))

	for _, p := range plans {
		allPlans = append(allPlans, toPlanRow(p))
	}

	return allPlans
}

func toPlanRow(p *adminv1.BillingPlan) *plan {
	return &plan{
		ID:                                  p.Id,
		Name:                                p.Name,
		DisplayName:                         p.DisplayName,
		Description:                         p.Description,
		TrialDays:                           strconv.Itoa(int(p.TrialPeriodDays)),
		Default:                             p.Default,
		Public:                              p.Public,
		QuotaNumProjects:                    p.Quotas.Projects,
		QuotaNumDeployments:                 p.Quotas.Deployments,
		QuotaNumSlotsTotal:                  p.Quotas.SlotsTotal,
		QuotaNumSlotsPerDeployment:          p.Quotas.SlotsPerDeployment,
		QuotaNumOutstandingInvites:          p.Quotas.OutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: p.Quotas.StorageLimitBytesPerDeployment,
	}
}

type plan struct {
	ID                                  string `header:"id" json:"id"`
	Name                                string `header:"name" json:"name"`
	DisplayName                         string `header:"display_name" json:"display_name"`
	Description                         string `header:"description" json:"description"`
	TrialDays                           string `header:"trial_days" json:"trial_days"`
	Default                             bool   `header:"default" json:"default"`
	Public                              bool   `header:"public" json:"public"`
	QuotaNumProjects                    string `header:"quota_num_projects" json:"quota_num_projects"`
	QuotaNumDeployments                 string `header:"quota_num_deployments" json:"quota_num_deployments"`
	QuotaNumSlotsTotal                  string `header:"quota_num_slots_total" json:"quota_num_slots_total"`
	QuotaNumSlotsPerDeployment          string `header:"quota_num_slots_per_deployment" json:"quota_num_slots_per_deployment"`
	QuotaNumOutstandingInvites          string `header:"quota_num_outstanding_invites" json:"quota_num_outstanding_invites"`
	QuotaStorageLimitBytesPerDeployment string `header:"quota_storage_limit_bytes_per_deployment" json:"quota_storage_limit_bytes_per_deployment"`
}

func (p *Printer) PrintMemberUsergroups(members []*adminv1.MemberUsergroup) {
	if len(members) == 0 {
		p.PrintfWarn("No user groups found\n")
		return
	}

	p.PrintData(toUsergroupsTable(members))
}

func toUsergroupsTable(members []*adminv1.MemberUsergroup) []*memberUsergroup {
	allMembers := make([]*memberUsergroup, 0, len(members))
	for _, ug := range members {
		allMembers = append(allMembers, toMemberUsergroupRows(ug))
	}
	return allMembers
}

func toMemberUsergroupRows(ug *adminv1.MemberUsergroup) *memberUsergroup {
	role := ug.RoleName
	if role == "" {
		role = "-"
	}
	return &memberUsergroup{
		Name:      ug.GroupName,
		Role:      role,
		CreatedOn: ug.CreatedOn.AsTime().Local().Format(time.DateTime),
		UpdatedOn: ug.UpdatedOn.AsTime().Local().Format(time.DateTime),
	}
}

type memberUsergroup struct {
	Name      string `header:"name" json:"name"`
	Role      string `header:"role" json:"role"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_at"`
	UpdatedOn string `header:"updated_on,timestamp(ms|utc|human)" json:"updated_at"`
}

func (p *Printer) PrintModelPartitions(partitions []*runtimev1.ModelPartition) {
	if len(partitions) == 0 {
		p.PrintfWarn("No partitions found\n")
		return
	}

	p.PrintData(toModelPartitionsTable(partitions))
}

func toModelPartitionsTable(partitions []*runtimev1.ModelPartition) []*modelPartition {
	res := make([]*modelPartition, 0, len(partitions))
	for _, s := range partitions {
		res = append(res, toModelPartitionRow(s))
	}
	return res
}

func toModelPartitionRow(s *runtimev1.ModelPartition) *modelPartition {
	data, err := json.Marshal(s.Data)
	if err != nil {
		panic(err)
	}

	var executedOn string
	if s.ExecutedOn != nil {
		executedOn = s.ExecutedOn.AsTime().Format(time.RFC3339)
	}

	return &modelPartition{
		Key:        s.Key,
		DataJSON:   string(data),
		ExecutedOn: executedOn,
		Elapsed:    (time.Duration(s.ElapsedMs) * time.Millisecond).String(),
		Error:      s.Error,
	}
}

type modelPartition struct {
	Key        string `header:"key" json:"key"`
	DataJSON   string `header:"data" json:"data"`
	ExecutedOn string `header:"executed_on,timestamp(ms|utc|human)" json:"executed_on"`
	Elapsed    string `header:"elapsed" json:"elapsed"`
	Error      string `header:"error" json:"error"`
}

func (p *Printer) PrintBillingIssues(errs []*adminv1.BillingIssue) {
	if len(errs) == 0 {
		return
	}

	p.PrintData(toBillingIssuesTable(errs))
}

func toBillingIssuesTable(errs []*adminv1.BillingIssue) []*billingIssue {
	res := make([]*billingIssue, 0, len(errs))
	for _, e := range errs {
		res = append(res, toBillingIssueRow(e))
	}
	return res
}

func toBillingIssueRow(e *adminv1.BillingIssue) *billingIssue {
	meta, err := json.Marshal(e.Metadata)
	if err != nil || !utf8.Valid(meta) {
		meta = []byte("{\"error\": \"failed to marshal metadata\"}")
	}
	return &billingIssue{
		Organization: e.Org,
		Type:         e.Type.String(),
		Level:        e.Level.String(),
		Metadata:     string(meta), // TODO pretty print
		EventTime:    e.EventTime.AsTime().Local().Format(time.DateTime),
	}
}

type billingIssue struct {
	Organization string `header:"organization" json:"organization"`
	Type         string `header:"type" json:"type"`
	Level        string `header:"level" json:"level"`
	Metadata     string `header:"metadata" json:"metadata"`
	EventTime    string `header:"event_time,timestamp(ms|utc|human)" json:"event_time"`
}

// PrintQueryResponse prints the query response in the desired format (human, json, csv)
func (p *Printer) PrintQueryResponse(res *runtimev1.QueryResolverResponse) {
	if len(res.Data) == 0 {
		p.PrintfWarn("No data found\n")
		return
	}

	switch p.Format {
	// Interceptor for human format
	case FormatHuman:
		headers := extractQueryHeaders(res.Schema)
		rows := make([][]string, len(res.Data))

		for i, row := range res.Data {
			rows[i] = make([]string, len(headers))
			for j, field := range headers {
				if val, ok := row.GetFields()[field]; ok {
					rows[i][j] = p.FormatValue(val.AsInterface())
				} else {
					rows[i][j] = "null"
				}
			}
		}

		tableprinter.New(p.dataOut()).Render(headers, rows, nil, false)
		return

	// Interceptor for CSV format
	case FormatCSV:
		headers := extractQueryHeaders(res.Schema)
		w := csv.NewWriter(p.dataOut())

		if err := w.Write(headers); err != nil {
			panic(fmt.Errorf("failed to write CSV headers: %w", err))
		}

		for _, row := range res.Data {
			record := make([]string, len(headers))
			for i, field := range headers {
				if val, ok := row.GetFields()[field]; ok {
					record[i] = p.FormatValue(val.AsInterface())
				} else {
					record[i] = ""
				}
			}
			if err := w.Write(record); err != nil {
				panic(fmt.Errorf("failed to write CSV row: %w", err))
			}
		}

		w.Flush()
		if err := w.Error(); err != nil {
			panic(fmt.Errorf("failed to flush CSV writer: %w", err))
		}

		return
	default:
		p.PrintData(res.Data)
	}
}

func extractQueryHeaders(schema *runtimev1.StructType) []string {
	headers := make([]string, len(schema.Fields))
	for i, field := range schema.Fields {
		headers[i] = field.Name
	}
	return headers
}
