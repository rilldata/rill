package printer

import (
	"path/filepath"
	"strings"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

func (p *Printer) PrintOrgs(orgs []*adminv1.Organization, defaultOrg string) {
	if len(orgs) == 0 {
		p.PrintlnWarn("No organizations found")
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
		CreatedAt: o.CreatedOn.AsTime().Format(time.DateTime),
	}
}

type organization struct {
	Name      string `header:"name" json:"name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}

func (p *Printer) PrintProjects(projs []*adminv1.Project) {
	if len(projs) == 0 {
		p.PrintlnWarn("No projects found")
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
		p.PrintlnWarn("No users found")
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

func (p *Printer) PrintMembers(members []*adminv1.Member) {
	if len(members) == 0 {
		p.PrintlnWarn("No members found")
		return
	}

	p.PrintData(toMemberTable(members))
}

func toMemberTable(members []*adminv1.Member) []*member {
	allMembers := make([]*member, 0, len(members))

	for _, m := range members {
		allMembers = append(allMembers, toMemberRow(m))
	}

	return allMembers
}

func toMemberRow(m *adminv1.Member) *member {
	return &member{
		Name:      m.UserName,
		Email:     m.UserEmail,
		RoleName:  m.RoleName,
		CreatedOn: m.CreatedOn.AsTime().Format(time.DateTime),
		UpdatedOn: m.UpdatedOn.AsTime().Format(time.DateTime),
	}
}

type member struct {
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
	p.PrintData(toInvitesTable(invites))
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
		CreatedAt: s.CreatedOn.AsTime().Format(time.DateTime),
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
		expiresOn = s.ExpiresOn.AsTime().Format(time.DateTime)
	}

	return &token{
		ID:        s.Id,
		CreatedOn: s.CreatedOn.AsTime().Format(time.DateTime),
		ExpiresOn: expiresOn,
	}
}

type token struct {
	ID        string `header:"id" json:"id"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	ExpiresOn string `header:"expires_on,timestamp(ms|utc|human)" json:"expires_on"`
}
