// Centered, responsive content container. Uses token-backed max-width + gutter
// so every page aligns to the same grid (design-system §8).

import type { ReactNode } from 'react';

export default function Container({
  children,
  className = '',
}: {
  children: ReactNode;
  className?: string;
}): JSX.Element {
  return <div className={`mx-auto w-full max-w-container px-gutter ${className}`}>{children}</div>;
}
