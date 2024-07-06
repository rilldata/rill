package printer

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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
	githubURL := o.GithubUrl
	if o.Subpath != "" {
		githubURL = filepath.Join(o.GithubUrl, "tree", o.ProdBranch, o.Subpath)
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

func (p *Printer) PrintMemberUsers(members []*adminv1.MemberUser) {
	if len(members) == 0 {
		p.PrintfWarn("No members found\n")
		return
	}

	p.PrintData(toMemberTable(members))
}

func toMemberTable(members []*adminv1.MemberUser) []*memberUser {
	allMembers := make([]*memberUser, 0, len(members))

	for _, m := range members {
		allMembers = append(allMembers, toMemberRow(m))
	}

	return allMembers
}

func toMemberRow(m *adminv1.MemberUser) *memberUser {
	return &memberUser{
		Name:      m.UserName,
		Email:     m.UserEmail,
		RoleName:  m.RoleName,
		CreatedOn: m.CreatedOn.AsTime().Local().Format(time.DateTime),
		UpdatedOn: m.UpdatedOn.AsTime().Local().Format(time.DateTime),
	}
}

type memberUser struct {
	Name      string `header:"name" json:"display_name"`
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	UpdatedOn string `header:"updated_on,timestamp(ms|utc|human)" json:"updated_on"`
}

func (p *Printer) PrintInvites(invites []*adminv1.UserInvite) {
	if len(invites) == 0 {
		return
	}
	p.PrintDataWithTitle(toInvitesTable(invites), "Invites pending acceptance")
}

func toInvitesTable(invites []*adminv1.UserInvite) []*userInvite {
	allInvites := make([]*userInvite, 0, len(invites))

	for _, i := range invites {
		allInvites = append(allInvites, toInviteRow(i))
	}
	return allInvites
}

func toInviteRow(i *adminv1.UserInvite) *userInvite {
	return &userInvite{
		Email:     i.Email,
		RoleName:  i.Role,
		InvitedBy: i.InvitedBy,
	}
}

type userInvite struct {
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	InvitedBy string `header:"invited_by" json:"invited_by"`
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
	return &service{
		Name:      s.Name,
		OrgName:   s.OrgName,
		CreatedAt: s.CreatedOn.AsTime().Local().Format(time.DateTime),
	}
}

type service struct {
	Name      string `header:"name" json:"name"`
	OrgName   string `header:"org_name" json:"org_name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}

func (p *Printer) PrintServiceTokens(sts []*adminv1.ServiceToken) {
	if len(sts) == 0 {
		return
	}
	p.PrintData(toServiceTokensTable(sts))
}

func toServiceTokensTable(tkns []*adminv1.ServiceToken) []*token {
	tokens := make([]*token, 0, len(tkns))

	for _, t := range tkns {
		tokens = append(tokens, toServiceTokenRow(t))
	}

	return tokens
}

func toServiceTokenRow(s *adminv1.ServiceToken) *token {
	var expiresOn string
	if !s.ExpiresOn.AsTime().IsZero() {
		expiresOn = s.ExpiresOn.AsTime().Local().Format(time.DateTime)
	}

	return &token{
		ID:        s.Id,
		CreatedOn: s.CreatedOn.AsTime().Local().Format(time.DateTime),
		ExpiresOn: expiresOn,
	}
}

type token struct {
	ID        string `header:"id" json:"id"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	ExpiresOn string `header:"expires_on,timestamp(ms|utc|human)" json:"expires_on"`
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
	expr := metricsview.NewExpressionFromProto(t.MetricsViewFilter)
	filter, err := metricsview.ExpressionToString(expr)
	if err != nil {
		panic(err)
	}

	row := &magicAuthToken{
		ID:        t.Id,
		Dashboard: t.MetricsView,
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
	Dashboard string `header:"dashboard" json:"dashboard"`
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
		PlanName:                     s.PlanName,
		PlanDisplayName:              s.PlanDisplayName,
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

func (p *Printer) PrintUsergroupMembers(members []*adminv1.MemberUser) {
	if len(members) == 0 {
		p.PrintfWarn("No members found\n")
		return
	}

	p.PrintData(toUsergroupMembersTable(members))
}

func toUsergroupMembersTable(members []*adminv1.MemberUser) []*usergroupMember {
	allMembers := make([]*usergroupMember, 0, len(members))

	for _, m := range members {
		allMembers = append(allMembers, toUsergroupMemberRow(m))
	}

	return allMembers
}

func toUsergroupMemberRow(m *adminv1.MemberUser) *usergroupMember {
	return &usergroupMember{
		Name:  m.UserName,
		Email: m.UserEmail,
	}
}

type usergroupMember struct {
	Name  string `header:"name" json:"name"`
	Email string `header:"email" json:"email"`
}
