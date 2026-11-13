// WeBond product wordmark + mark. Inline SVG (no third-party asset): a red
// rounded tile with a white trend line, wordmark in the header text colour.

export default function Logo({ className = '' }: { className?: string }): JSX.Element {
  return (
    <span className={`inline-flex items-center gap-2 ${className}`}>
      <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-brand text-text-inverse shadow-sm">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path
            d="M5 16l4-7 3.5 5 2.5-3.5L19 16"
            stroke="currentColor"
            strokeWidth="2.2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </span>
      <span className="text-xl font-bold tracking-tight">WeBond</span>
    </span>
  );
}
