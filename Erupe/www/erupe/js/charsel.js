var DoOnceActive = true;
function createNormalAlert(message) {
  parent.postMessage(message, "*");
}
function createGoodAlert(message) {
  parent.postMessage(message, "*");
}
function createErrorAlert(message) {
  parent.postMessage(message, "*");
}
function createCharListItem(name, uid, weapon, HR, GR, lastLogin, sex) {
	var icon;
	var active = "active";
	const unixTimestamp = lastLogin;
	const milliseconds = unixTimestamp * 1000;
	const dateObject = new Date(milliseconds);
	const humanDateFormat = dateObject.toLocaleString();
	dateObject.toLocaleString("en-US", {weekday: "long"});
	dateObject.toLocaleString("en-US", {month: "long"});
	dateObject.toLocaleString("en-US", {day: "numeric"});
	dateObject.toLocaleString("en-US", {year: "numeric"});
	dateObject.toLocaleString("en-US", {timeZoneName: "short"});
	lastLogin = humanDateFormat;
	lastLogin = lastLogin.split(' ')[0];
	if (sex == "M"){
		sex = "♂";
	}
	else{
		sex = "♀";
	}
	if (HR > 999){
		HR = 999;
	}
	if (GR > 999){
		GR = "999";
	}

	if (weapon == "片手剣"){
		weapon = "Sword & Shield";
		icon = "./ressources/icons/SS.png";
	}
	else if (weapon == "双剣"){
		weapon = "Dual Swords";
		icon = "./ressources/icons/DS.png";			
	}
	else if (weapon == "大剣"){
		weapon = "Great Sword";
		icon = "./ressources/icons/GS.png";
	}
	else if (weapon == "太刀"){
		weapon = "Long Sword";
		icon = "./ressources/icons/LS.png";
	}
	else if (weapon == "ハンマー"){
		weapon = "Hammer";
		icon = "./ressources/icons/H.png";
	}
	else if (weapon == "狩猟笛"){
		weapon = "Hunting Horn";
		icon = "./ressources/icons/HH.png";
	}
	else if (weapon == "ランス"){
		weapon = "Lance";
		icon = "./ressources/icons/L.png";		
	}
	else if (weapon == "ガンランス"){
		weapon = "Gunlance";
		icon = "./ressources/icons/GL.png";		
	}
	else if (weapon == "穿龍棍"){
		weapon = "Tonfa";
		icon = "./ressources/icons/T.png";		
	}
	else if (weapon == "スラッシュアックスF"){
		weapon = "Switch Axe F";
		icon = "./ressources/icons/SAF.png";		
	}
	else if (weapon == "マグネットスパイク"){
		weapon = "Magnet Spike";
		icon = "./ressources/icons/MS.png";		
	}
	else if (weapon == "ヘビィボウガン"){
		weapon = "Heavy Bowgun";
		icon = "./ressources/icons/HS.png";		
	}
	else if (weapon == "ライトボウガン"){
		weapon = "Light Bowgun";
		icon = "./ressources/icons/LB.png";		
	}
	else if (weapon == "弓"){
		weapon = "Bow";
		icon = "./ressources/icons/B.png";		
	}
	else{
	weapon = "Unknown"
		icon = "./ressources/icons/null.png";
	}
		
	if (DoOnceActive){
		DoOnceActive = false;
		var topDiv = $('<div/>')
		.attr("href", "#")
		.attr("uid", uid)
		.addClass("char-list-entry list-group-item list-group-item-action flex-column align-items-start active");
	}
	else{
		var topDiv = $('<div/>')
		.attr("href", "#")
		.attr("uid", uid)
		.addClass("char-list-entry list-group-item list-group-item-action flex-column align-items-start");
	}
	var topLine = $('<div/>')
	.addClass("Name_Player")
	.append($('<h1/>').addClass("mb-1").text(name)
	);
	var bottomLine = $('<div/>')
	.addClass("Info")
	.append($('<div id="icon_weapon"/>').prepend($('<img>',{id:'theImg',src:icon})))
	.append($('<div id="weapon_title"/>').text('Current Weapon'))
	.append($('<div id="weapon_name"/>').text(weapon))
	.append($('<div id="hr_lvl"/>').text('HR' + HR))
	.append($('<div id="gr_lvl"/>').text('GR' + GR))
	.append($('<div id="sex"/>').text(sex))
	.append($('<div id="uid"/>').text('ID: ' + uid))
	.append($('<div id="lastlogin"/>').text('LastLogin ' + lastLogin));
	topDiv.append(topLine);
	topDiv.append(bottomLine);
	$("#characterlist").append(topDiv);
}
$(function () {
	try {
		var charInfo = window.external.getCharacterInfo();
	} catch (e) {
		createErrorAlert("Error on getCharacterInfo!");
	}
	try {
		$xmlDoc = new ActiveXObject("Microsoft.XMLDOM");
		$xmlDoc.async = "false";
		$xmlDoc.loadXML(charInfo);
		$xml = $($xmlDoc);
	} catch (e) {
		createErrorAlert("Error parsing character info xml!" + e);
	}

	try {
		$($xml).find("Character").each(function () {
		  createCharListItem(
			$(this).attr('name'),
			$(this).attr('uid'),
			$(this).attr('weapon'),
			$(this).attr('HR'),
			$(this).attr('GR'),
			$(this).attr('lastLogin'),
			$(this).attr('sex')
			);
		});
	} catch (e) {
		createErrorAlert("Error searching character info xml!");
	}

	$(".char-list-entry").click(function () {
		if (!$(this).hasClass("active")) {
		  $(".char-list-entry.active").removeClass("active");
		  $(this).addClass("active");
		}
	});

$(function() {
    var selectedUid = $(".char-list-entry.active").attr("uid");
	$("#bt_new_char").on("click", function(e) {
		alert("NOT WORK");
	});
	$("#bt_delete_char").on("click", function(e) {
		alert("NOT WORK");
	});
});

$("#bt_confirm").on("click", function () {
	try{
		elementID = parent.document.getElementById("BlockGlobal");
		elementID.style.display = "block";
	} catch(e) {
		alert(e);
	}
    var selectedUid = $(".char-list-entry.active").attr("uid");
    try {
      window.external.selectCharacter(selectedUid, selectedUid)
    } catch (e) {
		createErrorAlert("Error on select character!");
	  	try{
			elementID = parent.document.getElementById("BlockGlobal");
			elementID.style.display = "none";
		} catch(e) {
			alert(e);
		}
    }
    setTimeout(function () {
      window.external.exitLauncher();
    }, 3000);
  });
});

// Enable to read JP text
function isKanji(ch) {
    return (ch >= "\u4e00" && ch <= "\u9faf") ||
	(ch >= "\u3400" && ch <= "\u4dbf");
}

function accumulativeParser(str, condition) {
    let accumulations = [];
    let accumulator = "";

    for (let i = 0; i < str.length; ++i) {
        let ch = str[i];

        if (condition(x)) {
            accumulator += ch;
        } else if (accumulator !== "") {
            accumulations.push(accumulator);
            accumulator = "";
        }
    }
    return accumulations;
}

function parseKanjiCompounds(str) {
    return accumulativeParser(str, isKanji);
}


