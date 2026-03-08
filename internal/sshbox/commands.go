package sshbox

import (
	"fmt"
	"strings"

	"portfolio-server/internal/data"
	"portfolio-server/internal/renderer"
)

func renderHelp() string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("HELP"))
	b.WriteString("\r\n")
	cmds := [][]string{
		{"  whoami        ", "Identity & tagline"},
		{"  experience    ", "Work history"},
		{"  projects      ", "Key projects & GitHub links"},
		{"  skills        ", "Full tech stack"},
		{"  stats         ", "Ratings & metrics"},
		{"  achievements  ", "Awards, hackathons"},
		{"  contact <msg> ", "Send a message via Discord webhook"},
		{"  clear         ", "Clear the screen"},
		{"  exit          ", "Close connection"},
	}
	for _, c := range cmds {
		b.WriteString(cyan + c[0] + reset + dim + c[1] + reset + "\r\n")
	}
	b.WriteString("\r\n")
	return b.String()
}

func renderWhoami(p *data.Profile) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("WHOAMI"))
	b.WriteString("\r\n")
	b.WriteString(bold + "  " + p.Name + reset + "\r\n")
	b.WriteString(cyan + "  " + p.Tagline + "\r\n" + reset)
	b.WriteString(dim + "  ─────────────────────────────────────────────────\r\n" + reset)
	b.WriteString(yellow + "  🎓  " + reset + p.Education.Degree + " @ " + p.Education.Institution + "  |  CGPA " + p.Education.CGPA + "  |  " + p.Education.Period + "\r\n")
	b.WriteString(yellow + "  📞  " + reset + p.Contact.Phone + "\r\n")
	b.WriteString(yellow + "  📧  " + reset + p.Contact.Email + "\r\n")
	b.WriteString(yellow + "  🔗  " + reset + p.Contact.LinkedIn + "\r\n")
	b.WriteString(yellow + "  🐙  " + reset + p.Contact.GitHub + "\r\n")
	b.WriteString(yellow + "  ⚡  " + reset + p.Contact.CodeForces + "\r\n")
	b.WriteString("\r\n")
	return b.String()
}

func renderExperience(exps []data.Experience) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("EXP"))
	b.WriteString("\r\n")

	for _, exp := range exps {
		b.WriteString(bold + "  " + exp.Role + reset + "  ·  " + cyan + exp.Company + reset + "  " + dim + "(" + exp.Location + " · " + exp.Period + ")" + reset + "\r\n")
		for _, h := range exp.Highlights {
			b.WriteString(dim + "  · " + reset + h + "\r\n")
		}
		b.WriteString("\r\n")
	}
	return b.String()
}

func renderProjects(projs []data.Project) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("PROJ"))
	b.WriteString("\r\n")

	for _, p := range projs {
		b.WriteString(bold + "  " + p.Name + reset + "\r\n")
		stackStr := strings.Join(p.RelatedSkills, " · ")
		b.WriteString(magenta + "  " + stackStr + reset + "\r\n")
		b.WriteString(dim + "  " + p.Description + reset + "\r\n")

		var links []string
		for name, url := range p.Links {
			links = append(links, fmt.Sprintf("%s: %s", name, url))
		}
		if len(links) > 0 {
			b.WriteString(magenta + "  " + strings.Join(links, " | ") + reset + "\r\n")
		}
		b.WriteString("\r\n")
	}
	return b.String()
}

func renderSkills(skills []data.SkillGroup) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("SKILLS"))
	b.WriteString("\r\n")

	for _, group := range skills {
		b.WriteString(yellow + fmt.Sprintf("  %-11s", group.Category) + reset)
		for i, item := range group.Items {
			b.WriteString(cyan + item + reset)
			if i < len(group.Items)-1 {
				b.WriteString(dim + "  ·  " + reset)
			}
		}
		b.WriteString("\r\n")
	}
	b.WriteString("\r\n")
	return b.String()
}

func renderStats(p *data.Profile) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("STATS"))
	b.WriteString("\r\n")
	for _, stat := range p.Stats {
		b.WriteString(yellow + "  " + fmt.Sprintf("%-20s", stat.Label) + reset + bold + stat.Value + reset + "\r\n")
	}
	b.WriteString("\r\n")
	return b.String()
}

func renderAchievements(p *data.Profile) string {
	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString(renderer.TextToASCII("WIN"))
	b.WriteString("\r\n")
	for _, a := range p.Achievements {
		b.WriteString("  " + cyan + "▸ " + reset + bold + a.Title + reset + "\r\n")
		b.WriteString(dim + "    " + a.Description + reset + "\r\n\r\n")
	}
	return b.String()
}
