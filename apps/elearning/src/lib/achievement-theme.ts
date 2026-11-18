import Award from '@lucide/svelte/icons/award';
import Bug from '@lucide/svelte/icons/bug';
import Crown from '@lucide/svelte/icons/crown';
import Flame from '@lucide/svelte/icons/flame';
import Gem from '@lucide/svelte/icons/gem';
import Gift from '@lucide/svelte/icons/gift';
import Heart from '@lucide/svelte/icons/heart';
import Medal from '@lucide/svelte/icons/medal';
import Rocket from '@lucide/svelte/icons/rocket';
import Star from '@lucide/svelte/icons/star';
import Target from '@lucide/svelte/icons/target';
import Trophy from '@lucide/svelte/icons/trophy';
import Zap from '@lucide/svelte/icons/zap';
import type { Component } from 'svelte';

export interface AchievementTheme {
  icon: Component<{ size?: number; class?: string }>;
  bg: string;
  glow: string;
  text: string;
  emoji: string;
}

// Palet per nama ikon dari katalog achievement; ikon tak dikenal memakai tema Trophy.
export const ACHIEVEMENT_THEMES: Record<string, AchievementTheme> = {
  Bug: {
    icon: Bug,
    bg: 'from-lime-400 via-green-400 to-emerald-500',
    glow: 'rgba(52,211,153,0.35)',
    text: 'from-lime-300 via-green-200 to-emerald-200',
    emoji: '🐛',
  },
  Target: {
    icon: Target,
    bg: 'from-sky-400 via-blue-500 to-indigo-500',
    glow: 'rgba(96,165,250,0.35)',
    text: 'from-sky-200 via-blue-200 to-indigo-200',
    emoji: '🎯',
  },
  Zap: {
    icon: Zap,
    bg: 'from-yellow-300 via-amber-400 to-orange-500',
    glow: 'rgba(251,191,36,0.35)',
    text: 'from-yellow-200 via-amber-200 to-orange-200',
    emoji: '⚡',
  },
  Trophy: {
    icon: Trophy,
    bg: 'from-amber-400 via-yellow-400 to-yellow-500',
    glow: 'rgba(251,191,36,0.4)',
    text: 'from-amber-200 via-yellow-200 to-amber-100',
    emoji: '🏆',
  },
  Flame: {
    icon: Flame,
    bg: 'from-orange-400 via-red-500 to-rose-600',
    glow: 'rgba(244,63,94,0.35)',
    text: 'from-orange-200 via-red-200 to-rose-200',
    emoji: '🔥',
  },
  Star: {
    icon: Star,
    bg: 'from-purple-400 via-violet-500 to-fuchsia-500',
    glow: 'rgba(167,139,250,0.35)',
    text: 'from-purple-200 via-violet-200 to-fuchsia-200',
    emoji: '⭐',
  },
  Award: {
    icon: Award,
    bg: 'from-teal-400 via-cyan-400 to-sky-500',
    glow: 'rgba(34,211,238,0.35)',
    text: 'from-teal-200 via-cyan-200 to-sky-200',
    emoji: '🎖️',
  },
  Crown: {
    icon: Crown,
    bg: 'from-yellow-400 via-amber-400 to-yellow-600',
    glow: 'rgba(251,191,36,0.45)',
    text: 'from-yellow-200 via-amber-200 to-yellow-100',
    emoji: '👑',
  },
  Heart: {
    icon: Heart,
    bg: 'from-pink-400 via-rose-500 to-red-500',
    glow: 'rgba(244,114,182,0.35)',
    text: 'from-pink-200 via-rose-200 to-red-200',
    emoji: '❤️',
  },
  Rocket: {
    icon: Rocket,
    bg: 'from-indigo-400 via-blue-500 to-violet-600',
    glow: 'rgba(99,102,241,0.35)',
    text: 'from-indigo-200 via-blue-200 to-violet-200',
    emoji: '🚀',
  },
  Medal: {
    icon: Medal,
    bg: 'from-orange-300 via-amber-400 to-yellow-500',
    glow: 'rgba(251,191,36,0.35)',
    text: 'from-orange-200 via-amber-200 to-yellow-200',
    emoji: '🥇',
  },
  Gift: {
    icon: Gift,
    bg: 'from-fuchsia-400 via-pink-500 to-rose-500',
    glow: 'rgba(232,121,249,0.35)',
    text: 'from-fuchsia-200 via-pink-200 to-rose-200',
    emoji: '🎁',
  },
  Gem: {
    icon: Gem,
    bg: 'from-cyan-400 via-teal-400 to-emerald-500',
    glow: 'rgba(34,211,238,0.35)',
    text: 'from-cyan-200 via-teal-200 to-emerald-200',
    emoji: '💎',
  },
};

export const themeFor = (icon: string | null | undefined): AchievementTheme =>
  (icon && ACHIEVEMENT_THEMES[icon]) || ACHIEVEMENT_THEMES.Trophy;
