var images = ['1.jpg', '2.jpg', '3.jpg', '4.jpg', '5.jpg', '7.png', '8.png', '9.png'];
var randomImage = Math.floor(Math.random() * images.length)
$(document).ready(function()
{
	$("body").css("background-image", "url('./ressources/Background/" + images[randomImage] + "')");
})