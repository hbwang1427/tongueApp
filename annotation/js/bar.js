 $(document).ready(function() {
   var scrollStart = ($(window).width() - $('.scrollbar').width()) / 2;
   $('.scroll').draggable({
     axis: 'x',
     containment: 'parent',
     drag: function() {
       var dragAmt = $('.scroll').position().left - scrollStart;
       var dragPercent = dragAmt * 0.168;
       $('.amount').text(Math.round(dragPercent) + '%');
     },
     stop: function() {
       var dragAmt = $('.scroll').position().left - (scrollStart + 696);
       $('.bar').animate({
         'margin-left': dragAmt
       }, 700, 'easeOutCubic');

     }
   });
 });
