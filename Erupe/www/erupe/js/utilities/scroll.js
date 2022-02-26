// Create New Variables
var time = 100; // Timer to react
var target;
var targetOffset;

// Button Top
$("#bt_up_char_list").on("click" ,function(e){
	e.preventDefault();
	target = $('.char-list-entry.active').prev('.char-list-entry');
	if (target.length == 0)
		target = $('.char-list-entry:last');
	scrollTo(target);
	$('.active').removeClass('active');
	target.addClass('active');
});

// Button Bottom
$("#bt_down_char_list").on("click" ,function(e){
	e.preventDefault();
	target = $('.char-list-entry.active').next('.char-list-entry');
	if (target.length == 0)
		target = $('.char-list-entry:first');
	scrollTo(target);
	$('.active').removeClass('active');
	target.addClass('active');
});

// Work Animation
function scrollTo(selector) {
	var offset = $(selector).offset();
    var $characterlist = $('#characterlist');
    $characterlist.animate({
		scrollTop: offset.top - ($characterlist.offset().top - $characterlist.scrollTop())
    }, time);
}





