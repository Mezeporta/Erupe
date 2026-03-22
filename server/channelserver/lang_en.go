package channelserver

func langEnglish() i18n {
	var i i18n

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

	i.guild.rookieGuildName = "Rookie Clan %d"
	i.guild.returnGuildName = "Return Clan %d"

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

	return i
}
