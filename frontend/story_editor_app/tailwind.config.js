/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.html", // Scan HTML files
    "./src/**/*.vue",  // Scan Vue SFCs for Tailwind classes
    "./src/**/*.js",   // Scan JS files (e.g., main.js if it contains classes)
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
