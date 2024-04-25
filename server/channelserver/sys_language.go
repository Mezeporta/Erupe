package channelserver

type Bead struct {
	id          int
	name        string
	description string
}
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
	diva struct {
		prayer struct {
			beads []Bead
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

		i.diva.prayer.beads = []Bead{
			{id: 1, name: "暴風の祈珠", description: "ーあらしまかぜのきじゅー\n暴風とは猛る思い。\n聞く者に勇気を与える。"},
			{id: 3, name: "断力の祈珠", description: "ーだんりきのきじゅー\n断力とは断ち切る思い。\n聴く者に新たな利からを授ける。"},
			{id: 4, name: "風韻の祈珠", description: "ーふういんのきじゅー\n風韻とは歌姫の艶。\n時々で異なる趣を醸し出す。"},
			{id: 8, name: "斬刃の祈珠", description: "ーざんばのきじゅー\n斬刃とはすべてを切り裂く力。\n集めるほどに声の透明感は増す。"},
			{id: 9, name: "打明の祈珠", description: "ーうちあかりのきじゅー\n打明とは熱い力。\n聴く者に活力を与える。"},
			{id: 10, name: "弾起の祈珠", description: "ーたまおこしのきじゅー\n弾起とは悠遠の記憶。\n聴く者に更なる力を授ける。"},
			{id: 11, name: "変続の祈珠", description: "ーへんぞくのきじゅー\n変続とは永久の言葉。\n聴く者に継続力を授ける。"},
			{id: 14, name: "万雷の祈珠", description: "ーばんらいのきじゅー\n万雷とは歌姫に集う民の意識。\n歌姫の声を伝播させる。"},
			{id: 15, name: "不動の祈珠", description: "ーうごかずのきじゅー\n不動とは圧力。聞く者に圧倒する力を与える。"},
			{id: 17, name: "結集の祈珠", description: "ーけっしゅうのきじゅー\n結集とは確固たる信頼。\n集めるほどに狩人たちの精神力となる。"},
			{id: 18, name: "歌護の祈珠", description: "ーうたまもりのきじゅー\n歌護とは歌姫の護り。\n集めるほどに狩人たちの支えとなる。"},
			{id: 19, name: "強撃の祈珠", description: "ーきょうげきのきじゅー\n強撃とは強い声色。\n聞く者の力を研ぎ澄ます。"},
			{id: 20, name: "封火の祈珠", description: "ーふうかのきじゅー\n封火とは火の封印。\n聞く者に耐性力を与える。"},
			{id: 21, name: "封水の祈珠", description: "ーふうすいのきじゅー\n封水とは水の封印。\n聞く者に耐性力を与える。"},
			{id: 22, name: "封氷の祈珠", description: "ーふうひょうのきじゅー\n封氷とは氷の封印。\n聞く者に耐性力を与える。"},
			{id: 23, name: "封雷の祈珠", description: "ーふうらいのきじゅー\n封雷とは雷の封印。\n聞く者に耐性力を与える。"},
			{id: 24, name: "封龍の祈珠", description: "ーふうりゅうのきじゅー\n封龍とは龍の封印。\n聞く者に耐性力を与える。"},
			{id: 25, name: "封属の祈珠", description: "ーふうぞくのきじゅー\n封続とは五つの属性の封印。\n聞く者に耐性力を与える。"},
		}

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
		i.diva.prayer.beads = []Bead{
			{id: 1, name: "Bead of Storms", description: "ーあらしまかぜのきじゅー\n暴風とは猛る思い。\n聞く者に勇気を与える。"},
			{id: 3, name: "Bead of Severing", description: "All damage types can sever tails\nPower to sever, inspire with might.\nEmpower those who hear, in new light."},
			{id: 4, name: "Bead of Vitality", description: "Increased red health recovery speed\nDiva's allure, a soothing balm.\nRenews one's vigor, with vitality and calm."},
			{id: 8, name: "Bead of Slashing", description: "Damage up for slashing weapons\nWith every slash, its voice rings out.\nGrowing ever sharper, without a doubt."},
			{id: 9, name: "Bead of Striking", description: "Damage up for striking weapons\nWith every blow, you strike with force.\nLet the power guide your course."},
			{id: 10, name: "Bead of Firing", description: "Damage up for shooting weapons\nA memory of might, empowering those who hear.\nBullet and body, soaring without fear."},
			{id: 11, name: "Bead of Tenacity", description: "ーへんぞくのきじゅー\n変続とは永久の言葉。\n聴く者に継続力を授ける。"},
			{id: 14, name: "Bead of Elements", description: "ーばんらいのきじゅー\n万雷とは歌姫に集う民の意識。\n歌姫の声を伝播させる。"},
			{id: 15, name: "Bead of Restraint", description: "ーうごかずのきじゅー\n不動とは圧力。聞く者に圧倒する力を与える。"},
			{id: 17, name: "Bead of Unity", description: "ーけっしゅうのきじゅー\n結集とは確固たる信頼。\n集めるほどに狩人たちの精神力となる。"},
			{id: 18, name: "Bead of Warding", description: "ーうたまもりのきじゅー\n歌護とは歌姫の護り。\n集めるほどに狩人たちの支えとなる。"},
			{id: 19, name: "Bead of Fury", description: "ーきょうげきのきじゅー\n強撃とは強い声色。\n聞く者の力を研ぎ澄ます。"},
			{id: 20, name: "Bead of Fireproof", description: "ーふうかのきじゅー\n封火とは火の封印。\n聞く者に耐性力を与える。"},
			{id: 21, name: "Bead of Waterproof", description: "ーふうすいのきじゅー\n封水とは水の封印。\n聞く者に耐性力を与える。"},
			{id: 22, name: "Bead of Iceproof", description: "ーふうひょうのきじゅー\n封氷とは氷の封印。\n聞く者に耐性力を与える。"},
			{id: 23, name: "Bead of Thunderproof", description: "ーふうらいのきじゅー\n封雷とは雷の封印。\n聞く者に耐性力を与える。"},
			{id: 24, name: "Bead of Dragonproof", description: "ーふうりゅうのきじゅー\n封龍とは龍の封印。\n聞く者に耐性力を与える。"},
			{id: 25, name: "Bead of Immunity", description: "ーふうぞくのきじゅー\n封続とは五つの属性の封印。\n聞く者に耐性力を与える。"},
		}
		i.timer = "Time: %02d:%02d:%02d.%03d (%df)"

		i.commands.noOp = "You don't have permission to use this command"
		i.commands.disabled = "%s command is disabled"
		i.commands.reload = "Reloading players..."
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
