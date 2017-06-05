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
			// encoded contents of ../rw-dir-powerbox-request.capnp:
			query: ['EAZQAQEAABEBF1EEAQH__N_F9TYo_t8AAAA'],
			//// Note that we could do this:
			// query: ['EAZQAQEAABEBF1EEAQH__OB5R1Q5MM4AAAA'],
			//// ..for a read-only directory. but for demo purposes
			//// it's convienent to be able to use this script in both
			//// grain types.
		}}, "*");
	});
});
