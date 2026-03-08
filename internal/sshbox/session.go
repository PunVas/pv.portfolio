package sshbox

import (
	"fmt"
	"io"
	"strings"
	"time"

	"portfolio-server/internal/data"
	"portfolio-server/internal/discord"
	"portfolio-server/internal/renderer"

	"golang.org/x/crypto/ssh"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	green   = "\033[1;32m"
	cyan    = "\033[1;96m"
	yellow  = "\033[1;33m"
	red     = "\033[1;31m"
	magenta = "\033[1;35m"
	dim     = "\033[2m"
)

func writeln(w io.Writer, s string) {
	fmt.Fprint(w, strings.ReplaceAll(s, "\n", "\r\n"))
}

func RunShell(ch ssh.Channel, store *data.Store, dc *discord.Client) {
	defer ch.Close()

	// ── 1. Profile image ──────────────────────────────────────────
	art, err := renderer.ImageToHalfBlock("assets/profile.jpg", 40)
	if err != nil {
		writeln(ch, cyan+"  [ add assets/profile.jpg to render your photo ]\r\n"+reset)
	} else {
		writeln(ch, art)
	}

	// ── 2. Boot sequence MOTD ─────────────────────────────────────
	time.Sleep(150 * time.Millisecond)
	bootLines := []struct{ color, text string }{
		{dim, ""},
		{green, "[init] POST check passed. RAM: ok. Imposter syndrome: suppressed........  OK"},
		{green, "[init] loading ~10k lines of Go. singleflight armed.................  OK"},
		{cyan, "[sys]  whoami............................... " + strings.ToLower(store.Profile.Name)},
		{cyan, "[sys]  caffeine level........................ CRITICALLY LOW"},
		{yellow, "[net]  renegotiating TLS. yes, again. blame OpenSSL.................  OK"},
		{green, "[db]   pulling work history (" + store.Profile.Experience[0].Location + ")..........  OK"},
		{green, fmt.Sprintf("[db]   indexing %d projects. none are todo apps. promise.............  OK", len(store.Profile.Projects))},
		{magenta, "[ai]   XLM-RoBERTa online. hindi NLI at 67.2% acc. mBERT shaking...  OK"},
		{magenta, "[perf] 9k req/sec backend up. P99 < 50ms. redis laughing at N+1s....  OK"},
		{green, "[sec]  no hardcoded secrets found. checked twice. you're welcome.....  OK"},
		{bold, "[boot] all systems nominal — type 'help' and let's talk."},
		{"", ""},
	}
	for _, line := range bootLines {
		writeln(ch, line.color+line.text+reset+"\r\n")
		time.Sleep(75 * time.Millisecond)
	}

	// ── 3. Name banner ────────────────────────────────────────────
	writeln(ch, "\r\n")
	puneetBanner := []string{
		" ██████╗ ██╗   ██╗███╗  ██╗███████╗███████╗████████╗",
		" ██╔══██╗██║   ██║████╗ ██║██╔════╝██╔════╝╚══██╔══╝",
		" ██████╔╝██║   ██║██╔██╗██║█████╗  █████╗     ██║   ",
		" ██╔═══╝ ██║   ██║██║╚████║██╔══╝  ██╔══╝     ██║   ",
		" ██║     ╚██████╔╝██║ ╚███║███████╗███████╗   ██║   ",
		" ╚═╝      ╚═════╝ ╚═╝  ╚══╝╚══════╝╚══════╝   ╚═╝   ",
	}
	for _, row := range puneetBanner {
		writeln(ch, cyan+bold+row+reset+"\r\n")
	}
	writeln(ch, cyan+"  "+store.Profile.Tagline+"  |  "+store.Profile.Education.Institution+"  |  "+store.Profile.Location+reset+"\r\n")
	writeln(ch, dim+"  ──────────────────────────────────────────────────────────"+reset+"\r\n\r\n")

	// ── 4. Interactive REPL ───────────────────────────────────────
	buf := make([]byte, 1)
	prompt := green + strings.Split(strings.ToLower(store.Profile.Name), " ")[0] + "@portfolio" + reset + ":" + cyan + "~" + reset + "$ "
	var cmd strings.Builder
	writeln(ch, prompt)

	for {
		cmd.Reset()
	reading:
		for {
			n, err := ch.Read(buf)
			if err != nil || n == 0 {
				return
			}
			b := buf[0]
			switch {
			case b == 3: // Ctrl-C
				writeln(ch, "^C\r\n"+prompt)
				cmd.Reset()
			case b == 4: // Ctrl-D
				writeln(ch, "\r\n"+dim+"Connection closed. Stay awesome.\r\n"+reset)
				return
			case b == 127 || b == 8: // Backspace
				if cmd.Len() > 0 {
					s := cmd.String()
					cmd.Reset()
					cmd.WriteString(s[:len(s)-1])
					fmt.Fprint(ch, "\b \b")
				}
			case b == '\r' || b == '\n':
				fmt.Fprint(ch, "\r\n")
				break reading
			default:
				if b >= 32 && b < 127 {
					cmd.WriteByte(b)
					fmt.Fprintf(ch, "%c", b)
				}
			}
		}

		input := strings.TrimSpace(cmd.String())
		if input == "" {
			writeln(ch, prompt)
			continue
		}

		if input == "exit" || input == "quit" || input == "logout" {
			writeln(ch, dim+"\r\n  Goodbye! Connection closed gracefully.\r\n"+reset)
			return
		}

		output := ProcessCommand(input, store, dc)
		writeln(ch, output)
		writeln(ch, prompt)
	}
}

func ProcessCommand(input string, store *data.Store, dc *discord.Client) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "help":
		return renderHelp()
	case "whoami":
		return renderWhoami(&store.Profile)
	case "experience":
		return renderExperience(store.Profile.Experience)
	case "projects":
		return renderProjects(store.Profile.Projects)
	case "skills":
		return renderSkills(store.Profile.Skills)
	case "stats":
		return renderStats(&store.Profile)
	case "achievements":
		return renderAchievements(&store.Profile)
	case "contact":
		if len(args) == 0 {
			return yellow + "  Usage: contact <your message>\r\n" + reset
		}
		msg := strings.Join(args, " ")
		dc.Send(msg, "Terminal visitor @ "+time.Now().Format("2006-01-02 15:04"))
		return green + "  ✓ Message sent! I'll reply soon.\r\n" + reset
	case "clear":
		return "\033[2J\033[H"
	default:
		return fmt.Sprintf(red+"  bash: %s: command not found  (try 'help')\r\n"+reset, command)
	}
}
