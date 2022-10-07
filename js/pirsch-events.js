(function () {
    "use strict";

    // respect Do-Not-Track
    if (navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    const script = document.querySelector("#pirscheventsjs");

    // include pages
    try {
        const include = script.getAttribute("data-include");
        const paths = include ? include.split(",") : [];

        if (paths.length) {
            let found = false;

            for (let i = 0; i < paths.length; i++) {
                if (new RegExp(paths[i]).test(location.pathname)) {
                    found = true;
                    break;
                }
            }

            if (!found) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
    }

    // exclude pages
    try {
        const exclude = script.getAttribute("data-exclude");
        const paths = exclude ? exclude.split(",") : [];

        for (let i = 0; i < paths.length; i++) {
            if (new RegExp(paths[i]).test(location.pathname)) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
    }

    // register event function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch/event";
    const disableQueryParams = script.hasAttribute("data-disable-query");
    const disableReferrer = script.hasAttribute("data-disable-referrer");
    const disableResolution = script.hasAttribute("data-disable-resolution");

    window.pirsch = function (name, options) {
        if (typeof name !== "string" || !name) {
            return Promise.reject("The event name for Pirsch is invalid (must be a non-empty string)! Usage: pirsch('event name', {duration: 42, meta: {key: 'value'}})");
        }

        return new Promise((resolve, reject) => {
            const meta = options && options.meta ? options.meta : {};

            for (let key in meta) {
                if (meta.hasOwnProperty(key)) {
                    meta[key] = String(meta[key]);
                }
            }

            if (navigator.sendBeacon(endpoint, JSON.stringify({
                url: disableQueryParams ? (location.href.includes('?') ? location.href.split('?')[0] : location.href).substring(0, 1800) : location.href.substring(0, 1800),
                title: document.title,
                referrer: (disableReferrer ? '' : document.referrer),
                screen_width: (disableResolution ? 0 : screen.width),
                screen_height: (disableResolution ? 0 : screen.height),
                event_name: name,
                event_duration: options && options.duration && typeof options.duration === "number" ? options.duration : 0,
                event_meta: meta
            }))) {
                resolve();
            } else {
                reject("error queuing event request");
            }
        });
    }
})();
