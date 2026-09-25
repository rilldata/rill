package ai

import (
	"context"
	"errors"
	"regexp"
	"slices"
)

var (
	// chatReferenceRegexp matches the references that the chat UI writes into prompts, e.g. <chat-reference>type="skill" skill="monthly-close"</chat-reference>.
	chatReferenceRegexp = regexp.MustCompile(`(?s)<chat-reference\b(.*?)</chat-reference>`)
	// chatReferenceAttrRegexp matches a key="value" pair in a chat reference.
	chatReferenceAttrRegexp = regexp.MustCompile(`(\w+)="([^"]*)"`)
)

// loadReferencedSkills pre-invokes the load_skill tool for each of the agent's skills that the user referenced in the prompt, in order of first reference.
// A skill in loaded is skipped: the model has it, and loading it again would repeat its whole body in the context.
// Tool errors are recorded in the session and don't fail the agent; only context cancellation is returned.
func loadReferencedSkills(ctx context.Context, s *Session, prompt string, skills []*Skill, agent string, loaded map[string]bool) error {
	for _, sk := range referencedSkills(prompt, skillsForAgent(skills, agent)) {
		if loaded[sk.Name] {
			continue
		}
		_, err := s.CallTool(ctx, RoleAssistant, LoadSkillName, nil, &LoadSkillArgs{Name: sk.Name})
		if err != nil && errors.Is(err, ctx.Err()) {
			return err
		}
	}
	return nil
}

// referencedSkills returns the skills referenced in a prompt with a chat reference of type "skill", once each and in order of first reference.
// References to skills that are not in the given list are ignored.
func referencedSkills(prompt string, skills []*Skill) []*Skill {
	var res []*Skill
	for _, ref := range chatReferenceRegexp.FindAllStringSubmatch(prompt, -1) {
		attrs := map[string]string{}
		for _, attr := range chatReferenceAttrRegexp.FindAllStringSubmatch(ref[1], -1) {
			attrs[attr[1]] = attr[2]
		}
		if attrs["type"] != "skill" {
			continue
		}

		idx := slices.IndexFunc(skills, func(sk *Skill) bool { return sk.Name == attrs["skill"] })
		if idx == -1 || slices.Contains(res, skills[idx]) {
			continue
		}
		res = append(res, skills[idx])
	}
	return res
}

// loadedSkills returns the names of the skills already loaded in the session, whether pre-invoked or called by the model.
// The predicates narrow the load_skill calls considered, e.g. to the calls of the current invocation when the model doesn't see earlier ones.
// A call whose result is an error, such as a skill that was not found yet, doesn't count: the model never got the body.
func loadedSkills(s *Session, predicates ...Predicate) map[string]bool {
	res := map[string]bool{}
	predicates = append(predicates, FilterByType(MessageTypeCall), FilterByTool(LoadSkillName))
	for _, call := range s.Messages(predicates...) {
		result, ok := s.Message(FilterByParent(call.ID), FilterByType(MessageTypeResult))
		if !ok || result.ContentType == MessageContentTypeError {
			continue
		}
		content, err := s.UnmarshalMessageContent(call)
		if err != nil {
			continue
		}
		if args, ok := content.(*LoadSkillArgs); ok {
			res[args.Name] = true
		}
	}
	return res
}
