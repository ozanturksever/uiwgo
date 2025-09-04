/** @type {import("tailwindcss").Config} */
module.exports = {
    content: [
        "!/vendor/**/*",
        "!/.devenv/**/*",
        "!/.cache/**/*",
        "!/node_modules/**/*",
        "!./examples/*/main.wasm",
        "!./**/*.md",
        "./examples/**/*.{html,js,go,jsx,tsx}",
        "./comps/**/*.{html,js,go,jsx,tsx}",
        "./dom/**/*.{html,js,go,jsx,tsx}",
        "./reactivity/**/*.{html,js,go,jsx,tsx}",
        "./router/**/*.{html,js,go,jsx,tsx}"
    ],
    theme: {
        extend: {}
    },
    plugins: []
};