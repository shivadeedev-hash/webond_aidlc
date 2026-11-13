// Seconds-left countdown used for "send a new code" buttons (resend cooldown 60 s).
import { useCallback, useEffect, useState } from 'react';

export function useCooldown(initialSeconds = 0): {
  secondsLeft: number;
  start: (seconds: number) => void;
} {
  const [secondsLeft, setSecondsLeft] = useState(initialSeconds);

  useEffect(() => {
    if (secondsLeft <= 0) return undefined;
    const id = window.setTimeout(() => setSecondsLeft((s) => s - 1), 1000);
    return () => window.clearTimeout(id);
  }, [secondsLeft]);

  const start = useCallback((seconds: number) => setSecondsLeft(seconds), []);
  return { secondsLeft, start };
}
