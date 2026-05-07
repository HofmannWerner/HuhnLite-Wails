const fs = require('fs');
const glob = require('glob');

const files = glob.sync('/Users/wernerhofmann/Projekte/HuhnLite/frontend/src/**/*.vue');

const inputRegex = /<q-input\s+([^>]*?)>/g;
const btnRegex = /<q-btn\s+([^>]*?)>/g;

files.forEach(file => {
    let content = fs.readFileSync(file, 'utf8');
    let original = content;

    // Process q-input
    content = content.replace(inputRegex, (match, attrs) => {
        // remove outlined
        attrs = attrs.replace(/\boutlined\b/g, '').trim();
        // ensure filled
        if (!/\bfilled\b/.test(attrs)) {
            attrs += ' filled';
        }
        // ensure stack-label
        if (!/\bstack-label\b/.test(attrs)) {
            attrs += ' stack-label';
        }
        // background color
        if (!/bg-color/.test(attrs) && !/:bg-color/.test(attrs)) {
            attrs += ` :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"`;
        }
        // cleanup extra spaces
        attrs = attrs.replace(/\s+/g, ' ').trim();
        return `<q-input ${attrs}>`;
    });

    // Process q-btn
    content = content.replace(btnRegex, (match, attrs) => {
        // user says: Alle Buttons sollen den Typ unelevated oder outline haben und zwingend das Attribut rounded nutzen
        if (!/\brounded\b/.test(attrs)) {
            attrs += ' rounded';
        }
        
        let hasOutline = /\boutline\b/.test(attrs);
        let hasUnelevated = /\bunelevated\b/.test(attrs);
        let hasFlat = /\bflat\b/.test(attrs);
        let hasClearBtn = /icon="close"\s+flat/.test(attrs) || /v-close-popup/.test(attrs) || /flat\s+round\s+dense\s+icon="edit"/.test(attrs) || /flat\s+round\s+dense\s+icon="delete"/.test(attrs) || /flat\s+dense\s+round\s+icon="edit"/.test(attrs) || /flat\s+dense\s+round\s+icon="delete"/.test(attrs) || /flat\s+dense\s+round\s+icon="menu"/.test(attrs);
        
        // Let's replace 'flat' with 'unelevated' if it's not a small icon button, or maybe keep flat for small icons?
        // Actually the user explicitly said "Alle Buttons sollen den Typ unelevated oder outline haben". Flat buttons should probably be changed to unelevated unless they are just icons. Let's make them unelevated or outline. 
        // If it doesnt have outline or unelevated, add unelevated, except maybe if it has flat
        if (!hasOutline && !hasUnelevated) {
            attrs = attrs.replace(/\bflat\b/g, '').trim();
            attrs += ' unelevated';
        }

        attrs = attrs.replace(/\s+/g, ' ').trim();
        return `<q-btn ${attrs}>`;
    });

    if (content !== original) {
        fs.writeFileSync(file, content);
        console.log(`Updated ${file}`);
    }
});
