(function() {
    "use strict";

    // respect Do-Not-Track
    if(navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    const script = document.querySelector("#pirscheventsjs");

    // exclude pages
    try {
        const exclude = script.getAttribute("data-exclude");
        const paths = exclude ? exclude.split(",") : [];

        for (let i = 0; i < paths.length; i++) {
            if (new RegExp(paths[i]).test(location.pathname)) {
                return;
            }
        }
    } catch(e) {
        console.error(e);
    }

    // register event function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch/event";
    window.pirsch = function(name, options) {
        if(typeof name !== "string" || !name) {
            return Promise.reject("The event name for Pirsch is invalid (must be a non-empty string)! Usage: pirsch('event name', {duration: 42, meta: {key: 'value'}})");
        }

        return new Promise((resolve, reject) => {
            const meta = options && options.meta ? options.meta : {};

            for(let key in meta) {
                if(meta.hasOwnProperty(key)) {
                    meta[key] = String(meta[key]);
                }
            }

            const req = new XMLHttpRequest();
            req.open("POST", endpoint);
            req.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
            req.onload = () => {
                if(req.status >= 200 && req.status < 300) {
                    resolve(req.response);
                } else {
                    reject(req.statusText);
                }
            };
            req.onerror = () => reject(req.statusText);
            req.send(JSON.stringify({
                url: location.href.substring(0, 1800),
                title: document.title,
                referrer: document.referrer,
                screen_width: screen.width,
                screen_height: screen.height,
                event_name: name,
                event_duration: options && options.duration && typeof options.duration === "number" ? options.duration : 0,
                event_meta: meta
            }));
        });
    }
})();
