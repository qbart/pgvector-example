/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['selector', '[data-mode="dark"]'],
  content: [
    "./ui/*.templ",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}

