function isIe6(){
	var ua = navigator.userAgent.toLowerCase();
	return /msie 6/.test(ua);
}
function $id(id){
	return document.getElementById(id);
}

function layoutTopLink(){
	if(layoutTopLink.timer!=null){
		window.clearTimeout(layoutTopLink.timer);
	}
	layoutTopLink.timer = window.setTimeout(function(){
		var el = $id("returnTop");
		//alert(document.body.clientHeight);
		var de = document.documentElement;
		if(de){
			el.style.right = "0px";
			//alert(de.scrollTop+de.clientHeight);
			el.style.top = (de.scrollTop+de.clientHeight-el.offsetHeight)+"px";
		}
	},50);
}
layoutTopLink.timer = null;

window.onload = function(){
	if(isIe6()){
		layoutTopLink();
		window.onscroll = function(){
			layoutTopLink();
		}
	}
}