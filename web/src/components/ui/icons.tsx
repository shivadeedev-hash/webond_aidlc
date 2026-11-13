// Inline SVG icon set (no emoji, license-free — design-system HARD RULE #3/#4).
// All icons inherit currentColor so they recolor via text-* token classes.

import type { SVGProps } from 'react';

type IconProps = Omit<SVGProps<SVGSVGElement>, 'width' | 'height'> & { size?: number };

function base({ size = 16, ...props }: IconProps): SVGProps<SVGSVGElement> {
  return {
    width: size,
    height: size,
    viewBox: '0 0 24 24',
    fill: 'none',
    'aria-hidden': true,
    ...props,
  };
}

/** Upward trend arrow — paired with success color + a sign for gains (never color alone). */
export function ArrowUp(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <path
        d="M12 19V5M12 5l-6 6M12 5l6 6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** Downward trend arrow — paired with danger color + a sign for losses. */
export function ArrowDown(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <path
        d="M12 5v14M12 19l6-6M12 19l-6-6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function CheckCircle(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
      <path
        d="M8.5 12.5l2.5 2.5 4.5-5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function Clock(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
      <path
        d="M12 7v5l3 2"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function XCircle(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
      <path d="M9 9l6 6M15 9l-6 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export function Ban(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
      <path d="M5.6 5.6l12.8 12.8" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export function AlertTriangle(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <path d="M12 3.5l9 16H3l9-16z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" />
      <path d="M12 10v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <circle cx="12" cy="17" r="1" fill="currentColor" />
    </svg>
  );
}

export function ChevronRight(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <path
        d="M9 6l6 6-6 6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function SearchIcon(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2" />
      <path d="M20 20l-3.5-3.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export function InfoIcon(props: IconProps): JSX.Element {
  return (
    <svg {...base(props)}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2" />
      <path d="M12 11v5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <circle cx="12" cy="7.5" r="1" fill="currentColor" />
    </svg>
  );
}
