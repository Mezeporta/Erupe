package channelserver

// Bead holds the display strings for a single kiju prayer bead type.
type Bead struct {
	ID          int
	Name        string
	Description string
}

type i18n struct {
	language string
	beads    []Bead
	cafe     struct {
		reset string
	}
	timer    string
	commands struct {
		noOp     string
		disabled string
		reload   string
		playtime string
		kqf      struct {
			get string
			set struct {
				error   string
				success string
			}
			version string
		}
		rights struct {
			error   string
			success string
		}
		course struct {
			error    string
			disabled string
			enabled  string
			locked   string
		}
		teleport struct {
			error   string
			success string
		}
		psn struct {
			error   string
			success string
			exists  string
		}
		discord struct {
			success string
		}
		ban struct {
			success string
			noUser  string
			invalid string
			error   string
			length  string
		}
		timer struct {
			enabled  string
			disabled string
		}
		ravi struct {
			noCommand string
			start     struct {
				success string
				error   string
			}
			multiplier string
			res        struct {
				success string
				error   string
			}
			sed struct {
				success string
			}
			request   string
			error     string
			noPlayers string
			version   string
		}
	}
	raviente struct {
		berserk        string
		extreme        string
		extremeLimited string
		berserkSmall   string
	}
	guild struct {
		rookieGuildName string
		returnGuildName string
		invite          struct {
			title   string
			body    string
			success struct {
				title string
				body  string
			}
			accepted struct {
				title string
				body  string
			}
			rejected struct {
				title string
				body  string
			}
			declined struct {
				title string
				body  string
			}
		}
	}
}

// beadName returns the localised name for a bead type.
func (i *i18n) beadName(beadType int) string {
	for _, b := range i.beads {
		if b.ID == beadType {
			return b.Name
		}
	}
	return ""
}

// beadDescription returns the localised description for a bead type.
func (i *i18n) beadDescription(beadType int) string {
	for _, b := range i.beads {
		if b.ID == beadType {
			return b.Description
		}
	}
	return ""
}

// getLangStrings returns the i18n string table for the configured language,
// falling back to English for unknown language codes.
func getLangStrings(s *Server) i18n {
	switch s.erupeConfig.Language {
	case "jp":
		return langJapanese()
	default:
		return langEnglish()
	}
}
