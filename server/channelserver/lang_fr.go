package channelserver

func langFrench() i18n {
	var i i18n

	i.language = "Français"
	i.cafe.reset = "Réinitialisation le %d/%d"
	i.timer = "Temps : %02d:%02d:%02d.%03d (%df)"

	i.commands.noOp = "Vous n'avez pas la permission d'utiliser cette commande"
	i.commands.disabled = "La commande %s est désactivée"
	i.commands.reload = "Rechargement des joueurs..."
	i.commands.playtime = "Temps de jeu : %d heure(s) %d minute(s) %d seconde(s)"

	i.commands.kqf.get = "KQF : %x"
	i.commands.kqf.set.error = "Erreur de commande. Format : %s set xxxxxxxxxxxxxxxx"
	i.commands.kqf.set.success = "KQF défini, veuillez changer de Zone/Monde"
	i.commands.kqf.version = "Cette commande est désactivée avant MHFG10"
	i.commands.rights.error = "Erreur de commande. Format : %s x"
	i.commands.rights.success = "Définir entier de droits : %d"
	i.commands.course.error = "Erreur de commande. Format : %s <nom>"
	i.commands.course.disabled = "Cours %s désactivé"
	i.commands.course.enabled = "Cours %s activé"
	i.commands.course.locked = "Le cours %s est verrouillé"
	i.commands.teleport.error = "Erreur de commande. Format : %s x y"
	i.commands.teleport.success = "Téléportation vers %d %d"
	i.commands.psn.error = "Erreur de commande. Format : %s <psn id>"
	i.commands.psn.success = "ID PSN connecté : %s"
	i.commands.psn.exists = "Cet ID PSN est déjà associé à un autre compte !"

	i.commands.discord.success = "Votre jeton Discord : %s"

	i.commands.ban.noUser = "Utilisateur introuvable"
	i.commands.ban.success = "%s a été banni avec succès"
	i.commands.ban.invalid = "ID de personnage invalide"
	i.commands.ban.error = "Erreur de commande. Format : %s <id> [durée]"
	i.commands.ban.length = " jusqu'au %s"

	i.commands.timer.enabled = "Minuteur de quête activé"
	i.commands.timer.disabled = "Minuteur de quête désactivé"

	i.commands.lang.usage = "Utilisation : %s <en|jp|fr|es>"
	i.commands.lang.invalid = "Langue inconnue %q. Prises en charge : en, jp, fr, es"
	i.commands.lang.success = "Langue définie sur %s"
	i.commands.lang.current = "Langue actuelle : %s"

	i.commands.ravi.noCommand = "Aucune commande Raviente spécifiée !"
	i.commands.ravi.start.success = "La Grande Chasse va commencer dans un instant"
	i.commands.ravi.start.error = "La Grande Chasse a déjà commencé !"
	i.commands.ravi.multiplier = "Le multiplicateur Raviente est actuellement de %.2fx"
	i.commands.ravi.res.success = "Envoi du soutien de résurrection !"
	i.commands.ravi.res.error = "Le soutien de résurrection n'a pas été demandé !"
	i.commands.ravi.sed.success = "Envoi du soutien de sédation si demandé !"
	i.commands.ravi.request = "Demande de soutien de sédation !"
	i.commands.ravi.error = "Commande Raviente non reconnue !"
	i.commands.ravi.noPlayers = "Personne n'a rejoint la Grande Chasse !"
	i.commands.ravi.version = "Cette commande est désactivée en dehors de MHFZZ"

	i.raviente.berserk = "<Grande Chasse : Frénésie> est en cours !"
	i.raviente.extreme = "<Grande Chasse : Extrême> est en cours !"
	i.raviente.extremeLimited = "<Grande Chasse : Extrême (Limitée)> est en cours !"
	i.raviente.berserkSmall = "<Grande Chasse : Frénésie (Réduite)> est en cours !"

	i.guild.rookieGuildName = "Clan Novice %d"
	i.guild.returnGuildName = "Clan Retour %d"

	i.guild.invite.title = "Invitation !"
	i.guild.invite.body = "Vous avez été invité à rejoindre\n「%s」\nSouhaitez-vous accepter ?"

	i.guild.invite.success.title = "Succès !"
	i.guild.invite.success.body = "Vous avez rejoint\n「%s」 avec succès."

	i.guild.invite.accepted.title = "Acceptée"
	i.guild.invite.accepted.body = "Le destinataire a accepté votre invitation à rejoindre\n「%s」."

	i.guild.invite.rejected.title = "Refusée"
	i.guild.invite.rejected.body = "Vous avez refusé l'invitation à rejoindre\n「%s」."

	i.guild.invite.declined.title = "Déclinée"
	i.guild.invite.declined.body = "Le destinataire a décliné votre invitation à rejoindre\n「%s」."

	i.beads = []Bead{
		{1, "Perle des Tempêtes", "Une perle de prière imprégnée du pouvoir des tempêtes.\nInvoque des vents déchaînés pour soutenir les alliés."},
		{3, "Perle de Tranchant", "Une perle de prière imprégnée du pouvoir tranchant.\nAccorde aux alliés une force de coupe accrue."},
		{4, "Perle de Vitalité", "Une perle de prière imprégnée de vitalité.\nAugmente les points de vie des alliés proches."},
		{8, "Perle de Guérison", "Une perle de prière imprégnée du pouvoir de guérison.\nProtège les alliés avec une énergie restauratrice."},
		{9, "Perle de Fureur", "Une perle de prière imprégnée d'énergie furieuse.\nEmbrasse les alliés d'une rage au combat."},
		{10, "Perle de Fléau", "Une perle de prière imprégnée de miasmes.\nInfuse les alliés d'un pouvoir venimeux."},
		{11, "Perle de Puissance", "Une perle de prière imprégnée d'une force brute.\nAccorde aux alliés une force accablante."},
		{14, "Perle du Tonnerre", "Une perle de prière imprégnée de foudre.\nCharge les alliés d'une force électrique."},
		{15, "Perle de Glace", "Une perle de prière imprégnée d'un froid glacial.\nAccorde aux alliés un pouvoir élémentaire glacé."},
		{17, "Perle de Feu", "Une perle de prière imprégnée d'une chaleur brûlante.\nEnflamme les alliés d'un pouvoir élémentaire ardent."},
		{18, "Perle d'Eau", "Une perle de prière imprégnée d'eau courante.\nAccorde aux alliés un pouvoir élémentaire aquatique."},
		{19, "Perle du Dragon", "Une perle de prière imprégnée d'énergie draconique.\nAccorde aux alliés un pouvoir élémentaire draconique."},
		{20, "Perle de Terre", "Une perle de prière imprégnée du pouvoir de la terre.\nAncre les alliés avec une force élémentaire tellurique."},
		{21, "Perle du Vent", "Une perle de prière imprégnée d'un vent rapide.\nAccorde aux alliés une agilité accrue."},
		{22, "Perle de Lumière", "Une perle de prière imprégnée d'une lumière radieuse.\nInspire les alliés avec une énergie lumineuse."},
		{23, "Perle d'Ombre", "Une perle de prière imprégnée d'obscurité.\nInfuse les alliés d'un pouvoir ténébreux."},
		{24, "Perle de Fer", "Une perle de prière imprégnée de la résistance du fer.\nAccorde aux alliés une défense renforcée."},
		{25, "Perle d'Immunité", "Une perle de prière imprégnée d'un pouvoir de scellement.\nAnnule les faiblesses élémentaires des alliés."},
	}

	return i
}
