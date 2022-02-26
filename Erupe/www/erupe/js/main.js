// Helper function to dynamically create Winsock alert box
function createWinsockAlert(message) {
	var tmpDiv = $('<div/>')
   	.attr("class", "WinsockAlertBoxID")
    .attr("role", "alert")
    .addClass("alert alert-info")    
    tmpDiv.append(message);  
	$("#alertBox").append(tmpDiv);
	scrollToBottom();
}

// Helper function to dynamically create normal alert box
function createNormalAlert(message) {
	var tmpDiv = $('<div/>')
   	.attr("class", "NormalAlertBoxID")
    .attr("role", "alert")
    .addClass("alert alert-secondary")    
    tmpDiv.append(message);  
	$("#alertBox").append(tmpDiv);
	scrollToBottom();
}

// Helper function to dynamically create good alert box
function createGoodAlert(message) {
	var tmpDiv = $('<div/>')
   	.attr("class", "GoodAlertBoxID")
    .attr("role", "alert")
    .addClass("alert alert-success")    
    tmpDiv.append(message);  
	$("#alertBox").append(tmpDiv);
	scrollToBottom();
}


function createErrorAlert(message) {
    var tmpDiv = $('<div/>')
	.attr("class", "ErrorAlertBoxID")
    .attr("role", "alert")
    .addClass("alert alert-danger")
    tmpDiv.append(message);  
	$("#alertBox").append(tmpDiv);
	scrollToBottom();
}

function scrollToBottom() {
	const messages = document.getElementById('alertBox');
	if (messages != null)
		messages.scrollTop = messages.scrollHeight;
}

$(function() {
    $("#configButton").click(function(){
        try{
            window.external.openMhlConfig();
        } catch(e){
            createErrorAlert("Error on openMhlConfig: " + e);
        }
    });
});

$(function() {
    $("#titlebar").on("click", function(e) {
        window.external.beginDrag(true);
    });
    $("#exit").on("click", function(e) {
        window.external.closeWindow();
    });
	$("#Reduce").on("click", function(e) {
        window.external.minimizeWindow();
	});
    $(window).on("message onmessage", function(e) {
        var data = e.originalEvent.data;
		CheckMessage(data);
    });
    doLauncherInitalize();
});

function CheckMessage(message){
	// Good Alert
	if (message == "Connected."){
		createGoodAlert(message);
	}
	// Normal Alert
	else if (message == "Authentification..."){
		createNormalAlert(message);	
	}
	else if (message == "After selecting a character, press [Start] button."){
		createNormalAlert(message);		
	}
	// Error Alert
	else {
		createErrorAlert(message);
	}
}

function doLauncherInitalize() {
	createWinsockAlert("Winsock Ver. [2.2]");
	createNormalAlert("After entering Erupe ID and Password, press [Connect] button.");
    try{
        window.external.getMhfMutexNumber();
    } catch(e){
        createErrorAlert("Error getting Mhf mutex number! " + e + ".");
    }

    try{
        var serverListXml = window.external.getServerListXml();
    } catch(e){
        createErrorAlert("Error getting serverlist.xml! " + e + ".");
    }

    if(serverListXml == ""){
        createErrorAlert("Got empty serverlist.xml!");
    }
    console.log(serverListXml);

    try{
        var lastServerIndex = window.external.getIniLastServerIndex();
    } catch(e){
        createErrorAlert("Error on getIniLastServerIndex: " + e + ".");
    }
    console.log("Last server index:" + lastServerIndex);

    try{
        window.external.setIniLastServerIndex(0);
    } catch(e){
        createErrorAlert("Error on setIniLastServerIndex: " + e + ".");
    }

    try{
        var mhfBootMode = window.external.getMhfBootMode();
    } catch(e){
        createErrorAlert("Error on getMhfBootMode: " + e + ".");
    }
    console.log("mhfBootMode:" + mhfBootMode);

    try{
        var userId = window.external.getUserId();
    } catch(e){
        createErrorAlert("Error on getUserId: " + e + ".");
    }
    console.log("userId:" + userId);

    try{
        var password = window.external.getPassword();
    } catch(e){
        createErrorAlert("Error on getPassword: " + e + ".");
    }
    console.log("password:" + password);
}	
	
	
	


