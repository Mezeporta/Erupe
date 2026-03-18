// protbot is a headless MHF protocol bot for testing Erupe server instances.
//
// Usage:
//
//	protbot --sign-addr 127.0.0.1:53312 --user test --pass test --action login
//	protbot --sign-addr 127.0.0.1:53312 --user test --pass test --action lobby
//	protbot --sign-addr 127.0.0.1:53312 --user test --pass test --action session
//	protbot --sign-addr 127.0.0.1:53312 --user test --pass test --action chat --message "Hello"
//	protbot --sign-addr 127.0.0.1:53312 --user test --pass test --action quests
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erupe-ce/cmd/protbot/scenario"
)

func main() {
	signAddr := flag.String("sign-addr", "127.0.0.1:53312", "Sign server address (host:port)")
	user := flag.String("user", "", "Username")
	pass := flag.String("pass", "", "Password")
	action := flag.String("action", "login", "Action to perform: login, lobby, session, chat, quests, achievement")
	message := flag.String("message", "", "Chat message to send (used with --action chat)")
	flag.Parse()

	if *user == "" || *pass == "" {
		fmt.Fprintln(os.Stderr, "error: --user and --pass are required")
		flag.Usage()
		os.Exit(1)
	}

	switch *action {
	case "login":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[done] Login successful!")
		_ = result.Channel.Close()

	case "lobby":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		if err := scenario.EnterLobby(result.Channel); err != nil {
			fmt.Fprintf(os.Stderr, "enter lobby failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		fmt.Println("[done] Lobby entry successful!")
		_ = result.Channel.Close()

	case "session":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		charID := result.Sign.CharIDs[0]
		if _, err := scenario.SetupSession(result.Channel, charID); err != nil {
			fmt.Fprintf(os.Stderr, "session setup failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		if err := scenario.EnterLobby(result.Channel); err != nil {
			fmt.Fprintf(os.Stderr, "enter lobby failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		fmt.Println("[session] Connected. Press Ctrl+C to disconnect.")
		waitForSignal()
		_ = scenario.Logout(result.Channel)

	case "chat":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		charID := result.Sign.CharIDs[0]
		if _, err := scenario.SetupSession(result.Channel, charID); err != nil {
			fmt.Fprintf(os.Stderr, "session setup failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		if err := scenario.EnterLobby(result.Channel); err != nil {
			fmt.Fprintf(os.Stderr, "enter lobby failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}

		// Register chat listener.
		scenario.ListenChat(result.Channel, func(msg scenario.ChatMessage) {
			fmt.Printf("[chat] <%s> (type=%d): %s\n", msg.SenderName, msg.ChatType, msg.Message)
		})

		// Send a message if provided.
		if *message != "" {
			if err := scenario.SendChat(result.Channel, 0x03, 1, *message, *user); err != nil {
				fmt.Fprintf(os.Stderr, "send chat failed: %v\n", err)
			}
		}

		fmt.Println("[chat] Listening for chat messages. Press Ctrl+C to disconnect.")
		waitForSignal()
		_ = scenario.Logout(result.Channel)

	case "quests":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		charID := result.Sign.CharIDs[0]
		if _, err := scenario.SetupSession(result.Channel, charID); err != nil {
			fmt.Fprintf(os.Stderr, "session setup failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		if err := scenario.EnterLobby(result.Channel); err != nil {
			fmt.Fprintf(os.Stderr, "enter lobby failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}

		data, err := scenario.EnumerateQuests(result.Channel, 0, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "enumerate quests failed: %v\n", err)
			_ = scenario.Logout(result.Channel)
			os.Exit(1)
		}
		fmt.Printf("[quests] Received %d bytes of quest data\n", len(data))
		_ = scenario.Logout(result.Channel)

	case "achievement":
		result, err := scenario.Login(*signAddr, *user, *pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		charID := result.Sign.CharIDs[0]
		if _, err := scenario.SetupSession(result.Channel, charID); err != nil {
			fmt.Fprintf(os.Stderr, "session setup failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}
		if err := scenario.EnterLobby(result.Channel); err != nil {
			fmt.Fprintf(os.Stderr, "enter lobby failed: %v\n", err)
			_ = result.Channel.Close()
			os.Exit(1)
		}

		// Step 1: Get current achievements.
		achs, err := scenario.GetAchievements(result.Channel, charID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get achievements failed: %v\n", err)
			_ = scenario.Logout(result.Channel)
			os.Exit(1)
		}
		fmt.Printf("[achievement] Total points: %d\n", achs.Points)
		hasNotify := false
		for _, e := range achs.Entries {
			marker := ""
			if e.Notify {
				marker = " ** RANK UP **"
				hasNotify = true
			}
			if e.Level > 0 || e.Notify {
				fmt.Printf("  [%2d] Level %d  Progress %d/%d  Trophy 0x%02X%s\n",
					e.ID, e.Level, e.Progress, e.Required, e.Trophy, marker)
			}
		}

		// Step 2: Mark as displayed if there were notifications.
		if hasNotify {
			fmt.Println("[achievement] Sending DISPLAYED_ACHIEVEMENT to acknowledge rank-ups...")
			if err := scenario.DisplayedAchievement(result.Channel); err != nil {
				fmt.Fprintf(os.Stderr, "displayed achievement failed: %v\n", err)
			}
			// Brief pause for fire-and-forget packet to be processed.
			<-time.After(500 * time.Millisecond)

			// Step 3: Re-fetch to verify notifications are cleared.
			achs2, err := scenario.GetAchievements(result.Channel, charID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "re-fetch achievements failed: %v\n", err)
			} else {
				cleared := true
				for _, e := range achs2.Entries {
					if e.Notify {
						fmt.Printf("  [%2d] STILL notifying (level %d) — not cleared!\n", e.ID, e.Level)
						cleared = false
					}
				}
				if cleared {
					fmt.Println("[achievement] All rank-up notifications cleared successfully.")
				}
			}
		} else {
			fmt.Println("[achievement] No pending rank-up notifications.")
		}

		_ = scenario.Logout(result.Channel)

	default:
		fmt.Fprintf(os.Stderr, "unknown action: %s (supported: login, lobby, session, chat, quests, achievement)\n", *action)
		os.Exit(1)
	}
}

// waitForSignal blocks until SIGINT or SIGTERM is received.
func waitForSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\n[signal] Shutting down...")
}
