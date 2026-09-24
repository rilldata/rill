package ai

import (
	"regexp"
	"slices"
)

var (
	// chatReferenceRegexp matches the references that the chat UI writes into prompts, e.g. <chat-reference>type="skill" skill="monthly-close"</chat-reference>.
	chatReferenceRegexp = regexp.MustCompile(`(?s)<chat-reference\b(.*?)</chat-reference>`)
	// chatReferenceAttrRegexp matches a key="value" pair in a chat reference.
	chatReferenceAttrRegexp = regexp.MustCompile(`(\w+)="([^"]*)"`)
)

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
// A call whose result is an error, such as a skill that was not found yet, doesn't count: the model never got the body.
func loadedSkills(s *Session) map[string]bool {
	res := map[string]bool{}
	for _, call := range s.Messages(FilterByType(MessageTypeCall), FilterByTool(LoadSkillName)) {
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
