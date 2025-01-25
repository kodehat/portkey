import Alpine from 'alpinejs';
import persist from '@alpinejs/persist';

import htmx from 'htmx.org';

// Work against 'unsafe-inline' CSP.
htmx.config.includeIndicatorStyles = false;
// Work against 'unsafe-eval' CSP.
htmx.config.selfRequestsOnly = true;
htmx.config.allowScriptTags = false;
htmx.config.allowEval = false;

window.Alpine = Alpine;

Alpine.plugin(persist);
Alpine.start();

console.log("JavaScript is running!");