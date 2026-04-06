package channelserver

func langChinese() i18n {
	var i i18n

	i.language = "中文"
	i.cafe.reset = "重置于 %d/%d"
	i.timer = "时间：%02d:%02d:%02d.%03d (%df)"

	i.commands.noOp = "您没有使用此命令的权限"
	i.commands.disabled = "%s 命令已禁用"
	i.commands.reload = "正在重新加载玩家..."
	i.commands.playtime = "游戏时间：%d 小时 %d 分钟 %d 秒"

	i.commands.kqf.get = "KQF：%x"
	i.commands.kqf.set.error = "命令错误。格式：%s set xxxxxxxxxxxxxxxx"
	i.commands.kqf.set.success = "已设置 KQF，请切换区域/世界"
	i.commands.kqf.version = "此命令在 MHFG10 之前已禁用"
	i.commands.rights.error = "命令错误。格式：%s x"
	i.commands.rights.success = "设置权限整数：%d"
	i.commands.course.error = "命令错误。格式：%s <名称>"
	i.commands.course.disabled = "%s 课程已禁用"
	i.commands.course.enabled = "%s 课程已启用"
	i.commands.course.locked = "%s 课程已锁定"
	i.commands.teleport.error = "命令错误。格式：%s x y"
	i.commands.teleport.success = "传送至 %d %d"
	i.commands.psn.error = "命令错误。格式：%s <psn id>"
	i.commands.psn.success = "已连接 PSN ID：%s"
	i.commands.psn.exists = "该 PSN ID 已连接到其他账户！"

	i.commands.discord.success = "您的 Discord 令牌：%s"

	i.commands.ban.noUser = "找不到用户"
	i.commands.ban.success = "已成功封禁 %s"
	i.commands.ban.invalid = "角色 ID 无效"
	i.commands.ban.error = "命令错误。格式：%s <id> [时长]"
	i.commands.ban.length = " 直到 %s"

	i.commands.timer.enabled = "任务计时器已启用"
	i.commands.timer.disabled = "任务计时器已禁用"

	i.commands.lang.usage = "用法：%s <en|jp|fr|es|zh>"
	i.commands.lang.invalid = "未知语言 %q。支持的语言：en, jp, fr, es, zh"
	i.commands.lang.success = "语言已设置为 %s"
	i.commands.lang.current = "当前语言：%s"

	i.commands.ravi.noCommand = "未指定 Raviente 命令！"
	i.commands.ravi.start.success = "大讨伐战即将开始"
	i.commands.ravi.start.error = "大讨伐战已经开始！"
	i.commands.ravi.multiplier = "Raviente 倍率当前为 %.2fx"
	i.commands.ravi.res.success = "正在发送复活支援！"
	i.commands.ravi.res.error = "尚未请求复活支援！"
	i.commands.ravi.sed.success = "若有请求则发送镇静支援！"
	i.commands.ravi.request = "请求镇静支援！"
	i.commands.ravi.error = "无法识别的 Raviente 命令！"
	i.commands.ravi.noPlayers = "无人参加大讨伐战！"
	i.commands.ravi.version = "此命令在 MHFZZ 以外已禁用"

	i.raviente.berserk = "<大讨伐战：狂暴> 正在进行！"
	i.raviente.extreme = "<大讨伐战：极限> 正在进行！"
	i.raviente.extremeLimited = "<大讨伐战：极限（限定）> 正在进行！"
	i.raviente.berserkSmall = "<大讨伐战：狂暴（小型）> 正在进行！"

	i.guild.rookieGuildName = "新人猎团 %d"
	i.guild.returnGuildName = "回归猎团 %d"

	i.guild.invite.title = "邀请！"
	i.guild.invite.body = "您已被邀请加入\n「%s」\n是否接受？"

	i.guild.invite.success.title = "成功！"
	i.guild.invite.success.body = "您已成功加入\n「%s」。"

	i.guild.invite.accepted.title = "已接受"
	i.guild.invite.accepted.body = "对方已接受您加入\n「%s」的邀请。"

	i.guild.invite.rejected.title = "已拒绝"
	i.guild.invite.rejected.body = "您拒绝了加入\n「%s」的邀请。"

	i.guild.invite.declined.title = "已婉拒"
	i.guild.invite.declined.body = "对方婉拒了您加入\n「%s」的邀请。"

	i.beads = []Bead{
		{1, "风暴之珠", "蕴含风暴之力的祈祷珠。\n召唤狂风助益同伴。"},
		{3, "斩击之珠", "蕴含斩击之力的祈祷珠。\n增强同伴的斩击力。"},
		{4, "活力之珠", "蕴含活力的祈祷珠。\n提升周围同伴的生命值。"},
		{8, "治愈之珠", "蕴含治愈之力的祈祷珠。\n以恢复能量守护同伴。"},
		{9, "狂怒之珠", "蕴含狂怒能量的祈祷珠。\n以战斗怒火激励同伴。"},
		{10, "瘴气之珠", "蕴含瘴气的祈祷珠。\n为同伴注入毒性之力。"},
		{11, "力量之珠", "蕴含原始力量的祈祷珠。\n赋予同伴压倒性的力量。"},
		{14, "雷鸣之珠", "蕴含闪电的祈祷珠。\n为同伴充填电力。"},
		{15, "寒冰之珠", "蕴含酷寒的祈祷珠。\n赋予同伴冰属性之力。"},
		{17, "烈火之珠", "蕴含灼热的祈祷珠。\n以烈火属性点燃同伴。"},
		{18, "流水之珠", "蕴含流水的祈祷珠。\n赋予同伴水属性之力。"},
		{19, "神龙之珠", "蕴含龙之能量的祈祷珠。\n赋予同伴龙属性之力。"},
		{20, "大地之珠", "蕴含大地之力的祈祷珠。\n以大地属性稳固同伴。"},
		{21, "疾风之珠", "蕴含疾风的祈祷珠。\n提升同伴的敏捷。"},
		{22, "光辉之珠", "蕴含光辉的祈祷珠。\n以光明能量鼓舞同伴。"},
		{23, "暗影之珠", "蕴含黑暗的祈祷珠。\n为同伴注入暗影之力。"},
		{24, "铁壁之珠", "蕴含钢铁之力的祈祷珠。\n为同伴强化防御。"},
		{25, "免疫之珠", "蕴含封印之力的祈祷珠。\n消除同伴的属性弱点。"},
	}

	return i
}
