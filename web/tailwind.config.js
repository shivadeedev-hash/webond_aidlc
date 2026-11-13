/** @type {import('tailwindcss').Config} */
// Colors reference CSS variables defined in src/styles/tokens.css.
// Components must use these token-backed classes — never hardcoded hex.
// tok() lets opacity modifiers (bg-primary/10, border-text-inverse/30) work
// with CSS-variable colours via color-mix.
const tok = (name) => `color-mix(in srgb, var(${name}) calc(<alpha-value> * 100%), transparent)`;

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // surfaces
        bg: tok('--color-bg'),
        surface: tok('--color-surface'),
        'surface-muted': tok('--color-surface-muted'),
        border: tok('--color-border'),
        // text
        text: tok('--color-text'),
        'text-muted': tok('--color-text-muted'),
        'text-inverse': tok('--color-text-inverse'),
        // brand
        header: tok('--color-header'),
        'header-text': tok('--color-header-text'),
        primary: {
          DEFAULT: tok('--color-primary'),
          hover: tok('--color-primary-hover'),
          active: tok('--color-primary-active'),
          soft: tok('--color-primary-soft'),
          'soft-strong': tok('--color-primary-soft-strong'),
        },
        accent: tok('--color-accent'),
        // semantic
        success: tok('--color-success'),
        danger: tok('--color-danger'),
        warning: tok('--color-warning'),
        info: tok('--color-info'),
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        DEFAULT: 'var(--radius-md)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        xl: 'var(--radius-xl)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        DEFAULT: 'var(--shadow-md)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
      },
      spacing: {
        'gutter': 'var(--space-gutter)',
        'section': 'var(--space-section)',
      },
      backgroundImage: {
        brand: 'var(--gradient-brand)',
      },
      maxWidth: {
        container: 'var(--container-max)',
      },
      transitionTimingFunction: {
        standard: 'var(--ease-standard)',
      },
      transitionDuration: {
        standard: 'var(--dur-standard)',
      },
    },
  },
  plugins: [],
};
