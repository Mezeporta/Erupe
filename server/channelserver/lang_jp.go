package channelserver

func langJapanese() i18n {
	var i i18n

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

	i.commands.playtime = "プレイ時間：%d時間%d分%d秒"

	i.commands.timer.enabled = "クエストタイマーが有効になりました"
	i.commands.timer.disabled = "クエストタイマーが無効になりました"

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

	i.guild.rookieGuildName = "新米猟団%d"
	i.guild.returnGuildName = "復帰猟団%d"

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

	return i
}
