package channelserver

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// onDiscordMessage handles receiving messages from discord and forwarding them ingame.
func (s *Server) onDiscordMessage(ds *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from our bot, or ones that are not in the correct channel.
	if m.Author.ID == ds.State.User.ID || m.ChannelID != s.erupeConfig.Discord.ChannelID {
		return
	}

	// Broadcast to the game clients.
	data := m.Content

	// Split on comma.
	result := strings.Split(data, " ")

	if result[0] == "!maintenancedate" && m.Author.ID == "836027554628370492" {
		replaceDays := dayConvert(result[1])
		replaceMonth := MonthConvert(result[3])

		s.BroadcastChatMessage("MAINTENANCE EXCEPTIONNELLE :")
		s.BroadcastChatMessage("Les serveurs seront temporairement inaccessibles le")
		s.BroadcastChatMessage(fmt.Sprintf("%s %s %s %s a partir de %s heures et %s minutes.", replaceDays, result[2], replaceMonth, result[4], result[5], result[6])) // Jour Mois Année Heures Minutes
		s.BroadcastChatMessage("Evitez de vous connecter durant cette periode. Veuillez nous")
		s.BroadcastChatMessage("excuser pour la gene occasionee. Merci de votre cooperation.")
		return
	} else if result[0] == "!maintenance" && m.Author.ID == "836027554628370492" {
		s.BroadcastChatMessage("RAPPEL DE MAINTENANCE DU MARDI (18H-22H): Les serveurs seront")
		s.BroadcastChatMessage("temporairement inaccessibles dans 15 minutes. Veuillez ne pas")
		s.BroadcastChatMessage("vous connecter ou deconnectez-vous maintenant, afin de ne pas")
		s.BroadcastChatMessage("perturber les operations de maintenance. Veuillez nous")
		s.BroadcastChatMessage("excuser pour la gene occasionnee. Merci de votre cooperation.")
		s.TimerUpdate(15, 0, true)
		return
	} else if result[0] == "!Rmaintenancedate" && m.Author.ID == "836027554628370492" {
		s.BroadcastChatMessage("RAPPEL DE MAINTENANCE EXCEPTIONNELLE: Les serveurs seront")
		s.BroadcastChatMessage("temporairement inaccessibles dans 15 minutes. Veuillez ne pas")
		s.BroadcastChatMessage("vous connecter ou deconnectez-vous maintenant, afin de ne pas")
		s.BroadcastChatMessage("perturber les operations de maintenance. Veuillez nous")
		s.BroadcastChatMessage("excuser pour la gene occasionnee. Merci de votre cooperation.")
		s.TimerUpdate(15, 1, true)
		return
	} else if result[0] == "!maintenanceStop" && m.Author.ID == "836027554628370492" {
		s.BroadcastChatMessage("INFORMATION: A titre exceptionnel, il n'y aura pas de")
		s.BroadcastChatMessage("maintenance de 18h a 22h, vous pouvez continuer de jouer")
		s.BroadcastChatMessage("librement jusqu'a la prochaine annonce de maintenance !")
		s.BroadcastChatMessage("Bonne chasse !")
		s.TimerUpdate(0, 0, false)
		return
	} else if result[0] == "!maintenanceStopExec" && m.Author.ID == "836027554628370492" {
		replaceDays := dayConvert(result[1])
		replaceMonth := MonthConvert(result[3])

		s.BroadcastChatMessage("INFORMATION: A titre exceptionnel, il n'y aura pas de")
		s.BroadcastChatMessage(fmt.Sprintf("maintenance le %s %s %s %s a partir de", replaceDays, result[2], replaceMonth, result[4])) // Jour Mois Année
		s.BroadcastChatMessage(fmt.Sprintf("%s heures et %s minutes, vous pouvez continuer de jouer", result[5], result[6]))           // Heures Minutes
		s.BroadcastChatMessage("librement jusqu'a la prochaine annonce de maintenance !")
		s.BroadcastChatMessage("Bonne chasse !")
		s.TimerUpdate(0, 1, false)
		return
	}

	message := fmt.Sprintf("[DISCORD] %s: %s", m.Author.Username, m.Content)
	s.BroadcastChatMessage(message)
}

func dayConvert(result string) string {
	var replaceDays string

	if result == "1" {
		replaceDays = "Lundi"
	} else if result == "2" {
		replaceDays = "Mardi"
	} else if result == "3" {
		replaceDays = "Mercredi"
	} else if result == "4" {
		replaceDays = "Jeudi"
	} else if result == "5" {
		replaceDays = "Vendredi"
	} else if result == "6" {
		replaceDays = "Samedi"
	} else if result == "7" {
		replaceDays = "Dimanche"
	} else {
		replaceDays = "NULL"
	}

	return replaceDays
}

func MonthConvert(result string) string {
	var replaceMonth string

	if result == "01" {
		replaceMonth = "Janvier"
	} else if result == "02" {
		replaceMonth = "Fevrier"
	} else if result == "03" {
		replaceMonth = "Mars"
	} else if result == "04" {
		replaceMonth = "Avril"
	} else if result == "05" {
		replaceMonth = "Mai"
	} else if result == "06" {
		replaceMonth = "Juin"
	} else if result == "07" {
		replaceMonth = "Juillet"
	} else if result == "08" {
		replaceMonth = "Aout"
	} else if result == "09" {
		replaceMonth = "Septembre"
	} else if result == "10" {
		replaceMonth = "Octobre"
	} else if result == "11" {
		replaceMonth = "Novembre"
	} else if result == "12" {
		replaceMonth = "Decembre"
	} else {
		replaceMonth = "NULL"
	}

	return replaceMonth
}

func (s *Server) TimerUpdate(timer int, typeStop int, disableAutoOff bool) {
	timertotal := 0
	for timer > 0 {
		time.Sleep(1 * time.Minute)
		timer -= 1
		timertotal += 1
		if disableAutoOff {
			// Un message s'affiche toutes les 10 minutes pour prévenir de la maintenance.
			if timertotal == 10 {
				timertotal = 0
				if typeStop == 0 {
					s.BroadcastChatMessage("RAPPEL DE MAINTENANCE DU MARDI (18H-22H): Les serveurs seront")
					s.BroadcastChatMessage(fmt.Sprintf("temporairement inaccessibles dans %d minutes. Veuillez ne pas", timer))
					s.BroadcastChatMessage("vous connecter ou deconnectez-vous maintenant, afin de ne pas")
					s.BroadcastChatMessage("perturber les operations de maintenance. Veuillez nous excuser")
					s.BroadcastChatMessage("pour la gene occasionnee. Merci de votre cooperation.")
				} else {
					s.BroadcastChatMessage("RAPPEL DE MAINTENANCE EXCEPTIONNELLE: Les serveurs seront")
					s.BroadcastChatMessage(fmt.Sprintf("temporairement inaccessibles dans %d minutes. Veuillez ne pas", timer))
					s.BroadcastChatMessage("vous connecter ou deconnectez-vous maintenant, afin de ne pas")
					s.BroadcastChatMessage("perturber les operations de maintenance. Veuillez nous excuser")
					s.BroadcastChatMessage("pour la gene occasionnee. Merci de votre cooperation.")
				}
			}
			// Déconnecter tous les joueurs du serveur.
			if timer == 0 {
				os.Exit(-1)
			}
		}
	}
}
