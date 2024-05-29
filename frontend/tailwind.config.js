/** @type {import('tailwindcss').Config} */
export default {

  content: ['./routes/**/*.{html,js,svelte,ts,tmpl,gotmpl,gohtml}', './assets/**/*.{html,js,svelte,ts,tmpl,gotmpl,gohtml}'],
  theme: {
    extend: {
      fontFamily: {
        'script': ['Rancho', 'ui-serif'],
        'sans': ['"Source Sans 3"', '"Merriweather Sans"', 'ui-sans-serif'],
        'serif': ['"Merriweather"', 'ui-serif'],
      }
    },
  },
  plugins: [],
}

