import Alpine from 'alpinejs';
import persist from '@alpinejs/persist';

import 'htmx.org';

window.Alpine = Alpine;

Alpine.plugin(persist);
Alpine.start();

console.log("JavaScript is running!");