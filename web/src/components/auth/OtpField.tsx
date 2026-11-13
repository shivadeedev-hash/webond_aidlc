// 6-digit one-time code input (US-002/004, Add-on 2026-11-12b). One native
// input (screen-reader and paste friendly) with the numeric keypad on phones and
// iOS/Android SMS-style autofill (autocomplete="one-time-code"). Non-digits are
// dropped so a pasted "123 456" still works.
import { TextField } from '../ui/Field';

export const OTP_LENGTH = 6;

export default function OtpField({
  label,
  value,
  onChange,
  hint,
  error,
  autoFocus,
}: {
  label: string;
  value: string;
  onChange: (code: string) => void;
  hint?: string;
  error?: string;
  autoFocus?: boolean;
}): JSX.Element {
  return (
    <TextField
      label={label}
      hint={hint}
      error={error}
      value={value}
      onChange={(e) => onChange(e.target.value.replace(/\D/g, '').slice(0, OTP_LENGTH))}
      inputMode="numeric"
      autoComplete="one-time-code"
      pattern="[0-9]{6}"
      maxLength={OTP_LENGTH + 2}
      placeholder="••••••"
      autoFocus={autoFocus}
      className="text-center font-mono text-2xl tracking-[0.5em]"
      required
    />
  );
}
