'use strict';

document.addEventListener('DOMContentLoaded', function() {
	document.getElementById('request-btn').addEventListener('click', function() {
		window.addEventListener('message', function(event) {
			if(event.data.rpcId !== 0) {
				return;
			}

			if(event.data.error) {
				console.error("rpc errored:", event.data.error);
				return;
			}

			var xhr = new XMLHttpRequest();
			xhr.onreadystatechange = function(event) {
				if(xhr.readyState === XMLHttpRequest.DONE) {
					window.location.reload(true);
				}
			}
			xhr.open("POST", "/filesystem-cap", true);
			xhr.overrideMimeType("text/plain; charset=x-user-defined");
			xhr.send(event.data.token);
		});
		window.parent.postMessage({powerboxRequest: {
			rpcId: 0,
			// encoded contents of ../ro-dir-powerbox-request.capnp:
			['EAZQAQEAABEBF1EEAQH__OB5R1Q5MM4AAAA'],
		}}, "*");
	});
});
