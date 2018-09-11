let menuIdx = 0;
let numItems = 0;
let menuUrl = '/';
$(document).ready(function() {
  numItems = $('.mainmenu').length;
  moveToMenuIdx(0);
});

function moveToMenuIdx(idx) {
  $('.mainmenu.table-primary').removeClass('table-primary');
  let item = $('.mainmenu').eq(idx);
  item.addClass('table-primary');
  menuIdx = idx;
  menuUrl = item.find('a')[0].href;
}

$(document).keypress(function(event) {
  let key = String.fromCharCode(event.which);
  if (key == "k" && menuIdx > 0) {
    moveToMenuIdx(menuIdx - 1);
  } else if (key == "j" && menuIdx < numItems-1) {
    moveToMenuIdx(menuIdx + 1);
  } else if (key == "l") {
    window.location = menuUrl;
  } else {
    return;
  }
  event.preventDefault();
});
