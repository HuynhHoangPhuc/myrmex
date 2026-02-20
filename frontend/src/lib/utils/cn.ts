import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// Combines clsx + tailwind-merge: resolves conditional classes and deduplicates Tailwind utilities
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}
