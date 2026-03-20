package channelserver

// Bead holds the display strings for a single kiju prayer bead type.
type Bead struct {
	ID          int
	Name        string
	Description string
}

// beadName returns the localised name for a bead type, falling back to a
// generic label if the type is not in the table.
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

type i18n struct {
	beads    []Bead
	language string
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
		invite struct {
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

func getLangStrings(s *Server) i18n {
	var i i18n
	switch s.erupeConfig.Language {
	case "jp":
		i.language = "日本語"
		i.cafe.reset = "%d/%dにリセット"
		i.timer = "タイマー：%02d'%02d\"%02d.%03d (%df)"

		i.commands.noOp = "You don't have permission to use this command"
		i.commands.disabled = "%sのコマンドは無効です"
		i.commands.reload = "リロードします"
		i.commands.kqf.get = "現在のキークエストフラグ：%x"
		i.commands.kqf.set.error = "キークエコマンドエラー　例：%s set xxxxxxxxxxxxxxxx"
		i.commands.kqf.set.success = "キークエストのフラグが更新されました。ワールド／ランドを移動してください"
		i.commands.kqf.version = "This command is disabled prior to MHFG10"
		i.commands.rights.error = "コース更新コマンドエラー　例：%s x"
		i.commands.rights.success = "コース情報を更新しました：%d"
		i.commands.course.error = "コース確認コマンドエラー　例：%s <name>"
		i.commands.course.disabled = "%sコースは無効です"
		i.commands.course.enabled = "%sコースは有効です"
		i.commands.course.locked = "%sコースはロックされています"
		i.commands.teleport.error = "テレポートコマンドエラー　構文：%s x y"
		i.commands.teleport.success = "%d %dにテレポート"
		i.commands.psn.error = "PSN連携コマンドエラー　例：%s <psn id>"
		i.commands.psn.success = "PSN「%s」が連携されています"
		i.commands.psn.exists = "PSNは既存のユーザに接続されています"

		i.commands.discord.success = "あなたのDiscordトークン：%s"

		i.commands.ban.noUser = "Could not find user"
		i.commands.ban.success = "Successfully banned %s"
		i.commands.ban.invalid = "Invalid Character ID"
		i.commands.ban.error = "Error in command. Format: %s <id> [length]"
		i.commands.ban.length = " until %s"

		i.commands.ravi.noCommand = "ラヴィコマンドが指定されていません"
		i.commands.ravi.start.success = "大討伐を開始します"
		i.commands.ravi.start.error = "大討伐は既に開催されています"
		i.commands.ravi.multiplier = "ラヴィダメージ倍率：ｘ%.2f"
		i.commands.ravi.res.success = "復活支援を実行します"
		i.commands.ravi.res.error = "復活支援は実行されませんでした"
		i.commands.ravi.sed.success = "鎮静支援を実行します"
		i.commands.ravi.request = "鎮静支援を要請します"
		i.commands.ravi.error = "ラヴィコマンドが認識されません"
		i.commands.ravi.noPlayers = "誰も大討伐に参加していません"
		i.commands.ravi.version = "This command is disabled outside of MHFZZ"

		i.raviente.berserk = "<大討伐：猛狂期>が開催されました！"
		i.raviente.extreme = "<大討伐：猛狂期【極】>が開催されました！"
		i.raviente.extremeLimited = "<大討伐：猛狂期【極】(制限付)>が開催されました！"
		i.raviente.berserkSmall = "<大討伐：猛狂期(小数)>が開催されました！"

		i.guild.invite.title = "猟団勧誘のご案内"
		i.guild.invite.body = "猟団「%s」からの勧誘通知です。\n「勧誘に返答」より、返答を行ってください。"

		i.guild.invite.success.title = "成功"
		i.guild.invite.success.body = "あなたは「%s」に参加できました。"

		i.guild.invite.accepted.title = "承諾されました"
		i.guild.invite.accepted.body = "招待した狩人が「%s」への招待を承諾しました。"

		i.guild.invite.rejected.title = "却下しました"
		i.guild.invite.rejected.body = "あなたは「%s」への参加を却下しました。"

		i.guild.invite.declined.title = "辞退しました"
		i.guild.invite.declined.body = "招待した狩人が「%s」への招待を辞退しました。"

		i.beads = []Bead{
			{1, "暴風の祈珠", "暴風の力を宿した祈珠。\n嵐を呼ぶ力で仲間を鼓舞する。"},
			{3, "断力の祈珠", "断力の力を宿した祈珠。\n斬撃の力を仲間に授ける。"},
			{4, "活力の祈珠", "活力の力を宿した祈珠。\n体力を高める力で仲間を鼓舞する。"},
			{8, "癒しの祈珠", "癒しの力を宿した祈珠。\n回復の力で仲間を守る。"},
			{9, "激昂の祈珠", "激昂の力を宿した祈珠。\n怒りの力を仲間に与える。"},
			{10, "瘴気の祈珠", "瘴気の力を宿した祈珠。\n毒霧の力を仲間に与える。"},
			{11, "剛力の祈珠", "剛力の力を宿した祈珠。\n強大な力を仲間に授ける。"},
			{14, "雷光の祈珠", "雷光の力を宿した祈珠。\n稲妻の力を仲間に与える。"},
			{15, "氷結の祈珠", "氷結の力を宿した祈珠。\n冷気の力を仲間に与える。"},
			{17, "炎熱の祈珠", "炎熱の力を宿した祈珠。\n炎の力を仲間に与える。"},
			{18, "水流の祈珠", "水流の力を宿した祈珠。\n水の力を仲間に与える。"},
			{19, "龍気の祈珠", "龍気の力を宿した祈珠。\n龍属性の力を仲間に与える。"},
			{20, "大地の祈珠", "大地の力を宿した祈珠。\n大地の力を仲間に与える。"},
			{21, "疾風の祈珠", "疾風の力を宿した祈珠。\n素早さを高める力を仲間に与える。"},
			{22, "光輝の祈珠", "光輝の力を宿した祈珠。\n光の力で仲間を鼓舞する。"},
			{23, "暗影の祈珠", "暗影の力を宿した祈珠。\n闇の力を仲間に与える。"},
			{24, "鋼鉄の祈珠", "鋼鉄の力を宿した祈珠。\n防御力を高める力を仲間に与える。"},
			{25, "封属の祈珠", "封属の力を宿した祈珠。\n属性を封じる力を仲間に与える。"},
		}
	default:
		i.language = "English"
		i.cafe.reset = "Resets on %d/%d"
		i.timer = "Time: %02d:%02d:%02d.%03d (%df)"

		i.commands.noOp = "You don't have permission to use this command"
		i.commands.disabled = "%s command is disabled"
		i.commands.reload = "Reloading players..."
		i.commands.playtime = "Playtime: %d hours %d minutes %d seconds"

		i.commands.kqf.get = "KQF: %x"
		i.commands.kqf.set.error = "Error in command. Format: %s set xxxxxxxxxxxxxxxx"
		i.commands.kqf.set.success = "KQF set, please switch Land/World"
		i.commands.kqf.version = "This command is disabled prior to MHFG10"
		i.commands.rights.error = "Error in command. Format: %s x"
		i.commands.rights.success = "Set rights integer: %d"
		i.commands.course.error = "Error in command. Format: %s <name>"
		i.commands.course.disabled = "%s Course disabled"
		i.commands.course.enabled = "%s Course enabled"
		i.commands.course.locked = "%s Course is locked"
		i.commands.teleport.error = "Error in command. Format: %s x y"
		i.commands.teleport.success = "Teleporting to %d %d"
		i.commands.psn.error = "Error in command. Format: %s <psn id>"
		i.commands.psn.success = "Connected PSN ID: %s"
		i.commands.psn.exists = "PSN ID is connected to another account!"

		i.commands.discord.success = "Your Discord token: %s"

		i.commands.ban.noUser = "Could not find user"
		i.commands.ban.success = "Successfully banned %s"
		i.commands.ban.invalid = "Invalid Character ID"
		i.commands.ban.error = "Error in command. Format: %s <id> [length]"
		i.commands.ban.length = " until %s"

		i.commands.timer.enabled = "Quest timer enabled"
		i.commands.timer.disabled = "Quest timer disabled"

		i.commands.ravi.noCommand = "No Raviente command specified!"
		i.commands.ravi.start.success = "The Great Slaying will begin in a moment"
		i.commands.ravi.start.error = "The Great Slaying has already begun!"
		i.commands.ravi.multiplier = "Raviente multiplier is currently %.2fx"
		i.commands.ravi.res.success = "Sending resurrection support!"
		i.commands.ravi.res.error = "Resurrection support has not been requested!"
		i.commands.ravi.sed.success = "Sending sedation support if requested!"
		i.commands.ravi.request = "Requesting sedation support!"
		i.commands.ravi.error = "Raviente command not recognised!"
		i.commands.ravi.noPlayers = "No one has joined the Great Slaying!"
		i.commands.ravi.version = "This command is disabled outside of MHFZZ"

		i.raviente.berserk = "<Great Slaying: Berserk> is being held!"
		i.raviente.extreme = "<Great Slaying: Extreme> is being held!"
		i.raviente.extremeLimited = "<Great Slaying: Extreme (Limited)> is being held!"
		i.raviente.berserkSmall = "<Great Slaying: Berserk (Small)> is being held!"

		i.guild.invite.title = "Invitation!"
		i.guild.invite.body = "You have been invited to join\n「%s」\nDo you want to accept?"

		i.guild.invite.success.title = "Success!"
		i.guild.invite.success.body = "You have successfully joined\n「%s」."

		i.guild.invite.accepted.title = "Accepted"
		i.guild.invite.accepted.body = "The recipient accepted your invitation to join\n「%s」."

		i.guild.invite.rejected.title = "Rejected"
		i.guild.invite.rejected.body = "You rejected the invitation to join\n「%s」."

		i.guild.invite.declined.title = "Declined"
		i.guild.invite.declined.body = "The recipient declined your invitation to join\n「%s」."

		i.beads = []Bead{
			{1, "Bead of Storms", "A prayer bead imbued with the power of storms.\nSummons raging winds to bolster allies."},
			{3, "Bead of Severing", "A prayer bead imbued with severing power.\nGrants allies increased cutting strength."},
			{4, "Bead of Vitality", "A prayer bead imbued with vitality.\nBoosts the health of those around it."},
			{8, "Bead of Healing", "A prayer bead imbued with healing power.\nProtects allies with restorative energy."},
			{9, "Bead of Fury", "A prayer bead imbued with furious energy.\nFuels allies with battle rage."},
			{10, "Bead of Blight", "A prayer bead imbued with miasma.\nInfuses allies with poisonous power."},
			{11, "Bead of Power", "A prayer bead imbued with raw might.\nGrants allies overwhelming strength."},
			{14, "Bead of Thunder", "A prayer bead imbued with lightning.\nCharges allies with electric force."},
			{15, "Bead of Ice", "A prayer bead imbued with freezing cold.\nGrants allies chilling elemental power."},
			{17, "Bead of Fire", "A prayer bead imbued with searing heat.\nIgnites allies with fiery elemental power."},
			{18, "Bead of Water", "A prayer bead imbued with flowing water.\nGrants allies water elemental power."},
			{19, "Bead of Dragon", "A prayer bead imbued with dragon energy.\nGrants allies dragon elemental power."},
			{20, "Bead of Earth", "A prayer bead imbued with earth power.\nGrounds allies with elemental earth force."},
			{21, "Bead of Wind", "A prayer bead imbued with swift wind.\nGrants allies increased agility."},
			{22, "Bead of Light", "A prayer bead imbued with radiant light.\nInspires allies with luminous energy."},
			{23, "Bead of Shadow", "A prayer bead imbued with darkness.\nInfuses allies with shadowy power."},
			{24, "Bead of Iron", "A prayer bead imbued with iron strength.\nGrants allies fortified defence."},
			{25, "Bead of Immunity", "A prayer bead imbued with sealing power.\nNullifies elemental weaknesses for allies."},
		}
	}
	return i
}
