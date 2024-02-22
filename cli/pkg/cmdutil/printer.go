package cmdutil

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

func Spinner(prefix string) *spinner.Spinner {
	// Other spinner options: https://github.com/briandowns/spinner#:~:text=90%20Character%20Sets.%20Some%20examples%20below%3A
	sp := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	sp.Prefix = prefix
	err := sp.Color("green", "bold")
	if err != nil {
		fmt.Println("invalid color and attribute list, Error: ", err)
	}

	return sp
}

func PrintUsers(p *printer.Printer, users []*adminv1.User) error {
	if len(users) == 0 {
		p.PrintlnWarn("No users found")
		return nil
	}

	err := p.PrintResource(toUsersTable(users))
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func PrintMembers(p *printer.Printer, members []*adminv1.Member) error {
	if len(members) == 0 {
		p.PrintlnWarn("No members found")
		return nil
	}

	err := p.PrintResource(toMemberTable(members))
	if err != nil {
		return err
	}

	return nil
}

func PrintInvites(p *printer.Printer, invites []*adminv1.UserInvite) error {
	if len(invites) == 0 {
		return nil
	}

	p.PrintlnSuccess("Pending user invites")
	err := p.PrintResource(toInvitesTable(invites))
	if err != nil {
		return err
	}

	return nil
}

func toUsersTable(users []*adminv1.User) []*user {
	allUsers := make([]*user, 0, len(users))

	for _, m := range users {
		allUsers = append(allUsers, toUserRow(m))
	}

	return allUsers
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

func toUserRow(m *adminv1.User) *user {
	return &user{
		Name:  m.DisplayName,
		Email: m.Email,
	}
}

type member struct {
	Name      string `header:"name" json:"display_name"`
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	UpdatedOn string `header:"updated_on,timestamp(ms|utc|human)" json:"updated_on"`
}

type user struct {
	Name  string `header:"name" json:"display_name"`
	Email string `header:"email" json:"email"`
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
