import typography from '@tailwindcss/typography';
import type { Config } from 'tailwindcss';
import tailwindcssAnimate from 'tailwindcss-animate';
import { fontFamily } from 'tailwindcss/defaultTheme';
import flattenColorPalette from 'tailwindcss/lib/util/flattenColorPalette';

// This plugin adds each Tailwind color as a global CSS variable, e.g. var(--gray-200).
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const addVariablesForColors = ({ addBase, theme }: any) => {
	const allColors = flattenColorPalette(theme('colors'));
	const newVars = Object.fromEntries(
		Object.entries(allColors).map(([key, val]) => [`--${key}`, val])
	);

	addBase({
		':root': newVars,
	});
};

const config: Config = {
	darkMode: ['class'],
	content: ['./src/**/*.{html,js,svelte,ts}'],
	safelist: ['dark'],
	theme: {
		container: {
			center: true,
			padding: '2rem',
			screens: {
				'2xl': '1400px',
			},
		},
		extend: {
			screens: {
				xs: '480px',
			},
			colors: {
				border: 'hsl(var(--border) / <alpha-value>)',
				input: 'hsl(var(--input) / <alpha-value>)',
				ring: 'hsl(var(--ring) / <alpha-value>)',
				background: 'hsl(var(--background) / <alpha-value>)',
				foreground: 'hsl(var(--foreground) / <alpha-value>)',
				primary: {
					DEFAULT: 'hsl(var(--primary) / <alpha-value>)',
					foreground: 'hsl(var(--primary-foreground) / <alpha-value>)',
				},
				secondary: {
					DEFAULT: 'hsl(var(--secondary) / <alpha-value>)',
					foreground: 'hsl(var(--secondary-foreground) / <alpha-value>)',
				},
				destructive: {
					DEFAULT: 'hsl(var(--destructive) / <alpha-value>)',
					foreground: 'hsl(var(--destructive-foreground) / <alpha-value>)',
				},
				muted: {
					DEFAULT: 'hsl(var(--muted) / <alpha-value>)',
					foreground: 'hsl(var(--muted-foreground) / <alpha-value>)',
				},
				accent: {
					DEFAULT: 'hsl(var(--accent) / <alpha-value>)',
					foreground: 'hsl(var(--accent-foreground) / <alpha-value>)',
				},
				popover: {
					DEFAULT: 'hsl(var(--popover) / <alpha-value>)',
					foreground: 'hsl(var(--popover-foreground) / <alpha-value>)',
				},
				card: {
					DEFAULT: 'hsl(var(--card) / <alpha-value>)',
					foreground: 'hsl(var(--card-foreground) / <alpha-value>)',
				},
				sidebar: {
					DEFAULT: 'hsl(var(--sidebar-background))',
					foreground: 'hsl(var(--sidebar-foreground))',
					primary: 'hsl(var(--sidebar-primary))',
					'primary-foreground': 'hsl(var(--sidebar-primary-foreground))',
					accent: 'hsl(var(--sidebar-accent))',
					'accent-foreground': 'hsl(var(--sidebar-accent-foreground))',
					border: 'hsl(var(--sidebar-border))',
					ring: 'hsl(var(--sidebar-ring))',
				},
				'secondary-darker': {
					DEFAULT: 'hsl(var(--secondary-darker) / <alpha-value>)',
					foreground: 'hsl(var(--secondary-darker-foreground) / <alpha-value>)',
				},
				// Portfolio-specific colors
				pink: {
					DEFAULT: 'hsl(0, 74%, 64%)',
					foreground: 'hsl(0, 74%, 64%)',
				},
				'color-1': 'hsl(var(--color-1))',
				'color-2': 'hsl(var(--color-2))',
				'color-3': 'hsl(var(--color-3))',
				'color-4': 'hsl(var(--color-4))',
				'color-5': 'hsl(var(--color-5))',
			},
			borderRadius: {
				xl: 'calc(var(--radius) + 4px)',
				lg: 'var(--radius)',
				md: 'calc(var(--radius) - 2px)',
				sm: 'calc(var(--radius) - 4px)',
			},
			fontFamily: {
				sans: [...fontFamily.sans],
				avant: ['AvantGarde', 'sans-serif'],
			},
			keyframes: {
				'accordion-down': {
					from: { height: '0' },
					to: { height: 'var(--bits-accordion-content-height)' },
				},
				'accordion-up': {
					from: { height: 'var(--bits-accordion-content-height)' },
					to: { height: '0' },
				},
				fadeIn: {
					from: { opacity: '0' },
					to: { opacity: '1' },
				},
				'caret-blink': {
					'0%,70%,100%': { opacity: '1' },
					'20%,50%': { opacity: '0' },
				},
				gradient: {
					to: {
						backgroundPosition: 'var(--bg-size) 0',
					},
				},
				rotate: {
					'0%': { transform: 'rotate(0deg) scale(10)' },
					'100%': { transform: 'rotate(-360deg) scale(10)' },
				},
				ping: {
					'75%, 100%': {
						transform: 'scale(2)',
						opacity: '0',
					},
				},
				rainbow: {
					'0%': { 'background-position': '0%' },
					'100%': { 'background-position': '200%' },
				},
				'spin-slow': {
					'0%': { transform: 'rotate(0deg)' },
					'100%': { transform: 'rotate(360deg)' },
				},
				bounce: {
					'0%, 100%': { transform: 'translateY(0) rotate(90deg)' },
					'50%': { transform: 'translateY(-5px) rotate(90deg)' },
				},
			},
			animation: {
				'accordion-down': 'accordion-down 0.2s ease-out',
				'accordion-up': 'accordion-up 0.2s ease-out',
				'caret-blink': 'caret-blink 1.25s ease-out infinite',
				'fast-pulse': 'pulse 1s cubic-bezier(0.4, 0, 0.6, 1) infinite',
				'spin-around': 'spin-around calc(var(--speed) * 2) infinite linear',
				slide: 'slide var(--speed) ease-in-out infinite alternate',
				shimmer: 'shimmer 8s infinite',
				gradient: 'gradient 8s linear infinite',
				rotate: 'rotate 5s linear infinite',
				ping: 'ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite',
				rainbow: 'rainbow var(--speed, 2s) infinite linear',
				'spin-slow': 'spin-slow 20s linear infinite',
				bounce: 'bounce 1.5s infinite',
			},
			backgroundImage: {
				'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
				'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
			},
		},
	},
	plugins: [typography, tailwindcssAnimate, addVariablesForColors],
};

export default config;
