package channelserver

func langSpanish() i18n {
	var i i18n

	i.language = "Español"
	i.cafe.reset = "Se reinicia el %d/%d"
	i.timer = "Tiempo: %02d:%02d:%02d.%03d (%df)"

	i.commands.noOp = "No tienes permiso para usar este comando"
	i.commands.disabled = "El comando %s está desactivado"
	i.commands.reload = "Recargando jugadores..."
	i.commands.playtime = "Tiempo de juego: %d hora(s) %d minuto(s) %d segundo(s)"

	i.commands.kqf.get = "KQF: %x"
	i.commands.kqf.set.error = "Error en el comando. Formato: %s set xxxxxxxxxxxxxxxx"
	i.commands.kqf.set.success = "KQF establecido, por favor cambia de Zona/Mundo"
	i.commands.kqf.version = "Este comando está desactivado antes de MHFG10"
	i.commands.rights.error = "Error en el comando. Formato: %s x"
	i.commands.rights.success = "Establecer entero de derechos: %d"
	i.commands.course.error = "Error en el comando. Formato: %s <nombre>"
	i.commands.course.disabled = "Curso %s desactivado"
	i.commands.course.enabled = "Curso %s activado"
	i.commands.course.locked = "El curso %s está bloqueado"
	i.commands.teleport.error = "Error en el comando. Formato: %s x y"
	i.commands.teleport.success = "Teletransportando a %d %d"
	i.commands.psn.error = "Error en el comando. Formato: %s <psn id>"
	i.commands.psn.success = "ID de PSN conectado: %s"
	i.commands.psn.exists = "Este ID de PSN ya está asociado a otra cuenta"

	i.commands.discord.success = "Tu token de Discord: %s"

	i.commands.ban.noUser = "No se encontró al usuario"
	i.commands.ban.success = "%s ha sido baneado con éxito"
	i.commands.ban.invalid = "ID de personaje inválido"
	i.commands.ban.error = "Error en el comando. Formato: %s <id> [duración]"
	i.commands.ban.length = " hasta el %s"

	i.commands.timer.enabled = "Temporizador de misión activado"
	i.commands.timer.disabled = "Temporizador de misión desactivado"

	i.commands.lang.usage = "Uso: %s <en|jp|fr|es|zh>"
	i.commands.lang.invalid = "Idioma desconocido %q. Compatibles: en, jp, fr, es, zh"
	i.commands.lang.success = "Idioma establecido en %s"
	i.commands.lang.current = "Idioma actual: %s"

	i.commands.ravi.noCommand = "No se especificó ningún comando de Raviente"
	i.commands.ravi.start.success = "La Gran Cacería comenzará en un momento"
	i.commands.ravi.start.error = "¡La Gran Cacería ya ha comenzado!"
	i.commands.ravi.multiplier = "El multiplicador de Raviente es actualmente %.2fx"
	i.commands.ravi.res.success = "¡Enviando apoyo de resurrección!"
	i.commands.ravi.res.error = "¡El apoyo de resurrección no ha sido solicitado!"
	i.commands.ravi.sed.success = "¡Enviando apoyo de sedación si fue solicitado!"
	i.commands.ravi.request = "¡Solicitando apoyo de sedación!"
	i.commands.ravi.error = "¡Comando de Raviente no reconocido!"
	i.commands.ravi.noPlayers = "¡Nadie se ha unido a la Gran Cacería!"
	i.commands.ravi.version = "Este comando está desactivado fuera de MHFZZ"

	i.raviente.berserk = "¡<Gran Cacería: Frenesí> está en curso!"
	i.raviente.extreme = "¡<Gran Cacería: Extremo> está en curso!"
	i.raviente.extremeLimited = "¡<Gran Cacería: Extremo (Limitado)> está en curso!"
	i.raviente.berserkSmall = "¡<Gran Cacería: Frenesí (Reducida)> está en curso!"

	i.guild.rookieGuildName = "Clan Novato %d"
	i.guild.returnGuildName = "Clan Regreso %d"

	i.guild.invite.title = "¡Invitación!"
	i.guild.invite.body = "Has sido invitado a unirte a\n「%s」\n¿Deseas aceptar?"

	i.guild.invite.success.title = "¡Éxito!"
	i.guild.invite.success.body = "Te has unido a\n「%s」 con éxito."

	i.guild.invite.accepted.title = "Aceptada"
	i.guild.invite.accepted.body = "El destinatario aceptó tu invitación para unirse a\n「%s」."

	i.guild.invite.rejected.title = "Rechazada"
	i.guild.invite.rejected.body = "Rechazaste la invitación para unirte a\n「%s」."

	i.guild.invite.declined.title = "Declinada"
	i.guild.invite.declined.body = "El destinatario declinó tu invitación para unirse a\n「%s」."

	i.beads = []Bead{
		{1, "Perla de Tormentas", "Una perla de oración imbuida con el poder de las tormentas.\nInvoca vientos furiosos para fortalecer a los aliados."},
		{3, "Perla de Corte", "Una perla de oración imbuida con poder cortante.\nOtorga a los aliados mayor fuerza de corte."},
		{4, "Perla de Vitalidad", "Una perla de oración imbuida con vitalidad.\nAumenta los puntos de vida de los aliados cercanos."},
		{8, "Perla de Curación", "Una perla de oración imbuida con poder curativo.\nProtege a los aliados con energía restauradora."},
		{9, "Perla de Furia", "Una perla de oración imbuida con energía furiosa.\nImbuye a los aliados con rabia de combate."},
		{10, "Perla de Plaga", "Una perla de oración imbuida con miasma.\nInfunde a los aliados con poder venenoso."},
		{11, "Perla de Poder", "Una perla de oración imbuida con fuerza bruta.\nOtorga a los aliados una fuerza abrumadora."},
		{14, "Perla del Trueno", "Una perla de oración imbuida con rayos.\nCarga a los aliados con fuerza eléctrica."},
		{15, "Perla de Hielo", "Una perla de oración imbuida con frío glacial.\nOtorga a los aliados poder elemental helado."},
		{17, "Perla de Fuego", "Una perla de oración imbuida con calor abrasador.\nEnciende a los aliados con poder elemental ígneo."},
		{18, "Perla de Agua", "Una perla de oración imbuida con agua fluyente.\nOtorga a los aliados poder elemental acuático."},
		{19, "Perla del Dragón", "Una perla de oración imbuida con energía dracónica.\nOtorga a los aliados poder elemental dracónico."},
		{20, "Perla de Tierra", "Una perla de oración imbuida con el poder de la tierra.\nAfianza a los aliados con fuerza elemental telúrica."},
		{21, "Perla del Viento", "Una perla de oración imbuida con viento veloz.\nOtorga a los aliados mayor agilidad."},
		{22, "Perla de Luz", "Una perla de oración imbuida con luz radiante.\nInspira a los aliados con energía luminosa."},
		{23, "Perla de Sombra", "Una perla de oración imbuida con oscuridad.\nInfunde a los aliados con poder sombrío."},
		{24, "Perla de Hierro", "Una perla de oración imbuida con la resistencia del hierro.\nOtorga a los aliados una defensa reforzada."},
		{25, "Perla de Inmunidad", "Una perla de oración imbuida con poder de sellado.\nAnula las debilidades elementales de los aliados."},
	}

	return i
}
