package channelserver

type i18n struct {
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
	}
	return i
}
