// Returns the scale factor required to get the image dimensions to fit into the given window
// dimensions.
function getScaleFactor(img_w, img_h, window_w, window_h) {
    let scale_factor = window_w / img_w;
    return (scale_factor * img_h > window_h) ? window_h / img_h : scale_factor;
}
// Scale down images that are wider than the page so they fit on the page.
// Waits for naturalHeight/Width to become available by using a jQuery plugin.
$(document).imagesLoaded().progress(function(instance, image) {
    let el = image.img;
    if (!image.isLoaded) {
        console.log("Failed to load image: " + el.src);
    $(el).closest('div.feeditem').remove();
        return;
    }
    let el_w = el.naturalWidth || el.videoWidth || el.width;
    let el_h = el.naturalHeight || el.videoHeight || el.height;
    if (!el_w || !el_h) {
        console.log("Error getting width/height for image: " + el.src);
        return;
    }
    let max_w = window.innerWidth - el.x - 50;
    let max_h = window.innerHeight - 100;
    let scale_factor = getScaleFactor(el_w, el_h, max_w, max_h)
    let new_w = Math.floor(scale_factor * el_w);
    let new_h = Math.floor(scale_factor * el_h);
    /*
        console.log(
          "Processing image: "+
          "  NW/NH: " + el_w + "," + el_h + " (" + (el_w/el_h).toFixed(4) + ")" +
          "  Window W/H: " + max_w + "," + max_h + " (" + (max_w/max_h).toFixed(4) + ")" +
          "  Scalefactor: " + scale_factor.toFixed(4) +
          "  New W/H: " + new_w + "," + new_h + " (" + (new_w/new_h).toFixed(4) + ")" +
          "  "+el.src
        );
    */
    el.style.width = new_w + "px";
    el.style.height = new_h + "px";
});
$(document).ready(function() {
    let max_height = window.innerHeight - 100;
    $('.videocontainer').each(function(idx, el) {
        el.style.maxHeight = max_height + "px";
    });
});

function scrollToNextItem(is_up) {
   let window_top_y = $(window).scrollTop();
   let new_top_y = 0;

   let items = [0, window_top_y];
   $(".feeditem").each(function(idx, el) {
       items.push(Math.floor($(el).offset().top));
   })
   if (is_up) {
      items.sort((a, b) => b-a);    // Descending order
      new_top_y = items.find((val) => val < window_top_y) || 0;
   } else {
      items.sort((a, b) => a-b);    // Ascending order
      new_top_y = items.find((val) => val > window_top_y) || items.pop();
   }
   window.scrollTo(0, new_top_y);
}
$(document).keypress(function(event) {
    let key = String.fromCharCode(event.which);
    if (key == "k" || key == "j") {               // Up/Down
        scrollToNextItem(key == "k")

    } else if (key == "h" && globals.currentPageNum > 1) {        // Previous page
        window.location = globals.previousPageLink;

    } else if (key == "l" && globals.currentPageNum < globals.numPages) {        // Next page
        window.location = globals.nextPageLink;

    } else if (key == "i") {        // Home
        window.location = '/';
    } else {
        return;
    }
    event.preventDefault();
});
