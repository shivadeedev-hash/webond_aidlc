// Email verification (US-002, Add-on 2026-11-12b). The user types the 6-digit
// code from the e-mail (works when the mail is read on another device), can
// send a new code after 60 s, and then logs in.

import { useState, type FormEvent } from 'react';
import { Link, useLocation } from 'react-router-dom';
import Container from '../components/layout/Container';
import AuthShell from '../components/auth/AuthShell';
import OtpField, { OTP_LENGTH } from '../components/auth/OtpField';
import { TextField } from '../components/ui/Field';
import Button from '../components/ui/Button';
import { authService } from '../services/auth';
import { useI18n } from '../i18n/useI18n';
import { useCooldown } from '../hooks/useCooldown';
import { ApiClientError } from '../lib/errors';
import { CheckCircle } from '../components/ui/icons';

export const RESEND_COOLDOWN_SECONDS = 60;

export default function VerifyEmail(): JSX.Element {
  const { t } = useI18n();
  const location = useLocation();
  const stateEmail = (location.state as { email?: string } | null)?.email ?? '';

  const [email, setEmail] = useState(stateEmail);
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);
  // A code was just sent by registration, so start with the cooldown running.
  const { secondsLeft, start } = useCooldown(stateEmail ? RESEND_COOLDOWN_SECONDS : 0);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);
    setNotice(null);
    setSubmitting(true);
    try {
      await authService.verify({ email: email.trim(), code });
      setDone(true);
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleResend(): Promise<void> {
    setError(null);
    setNotice(null);
    try {
      await authService.resendVerification({ email: email.trim() });
      setCode('');
      setNotice(t('auth.codeResent'));
      start(RESEND_COOLDOWN_SECONDS);
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    }
  }

  return (
    <Container className="py-section">
      <AuthShell>
        <h1 className="text-2xl font-bold">{t('auth.verifyEmail')}</h1>

        {done ? (
          <div className="mt-6">
            <p className="inline-flex items-center gap-2 font-medium text-success" role="status">
              <CheckCircle size={18} />
              {t('auth.verifySuccess')}
            </p>
            <Link to="/login" className="mt-4 block">
              <Button variant="primary" className="w-full">
                {t('auth.login')}
              </Button>
            </Link>
          </div>
        ) : (
          <>
            <p className="mt-2 text-sm text-text-muted">{t('auth.checkEmail')}</p>

            <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
              <TextField
                label={t('auth.email')}
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              <OtpField
                label={t('auth.verifyCode')}
                value={code}
                onChange={setCode}
                autoFocus={!!stateEmail}
              />
              {error && (
                <p role="alert" className="text-sm text-danger">
                  {error}
                </p>
              )}
              {notice && (
                <p role="status" className="text-sm text-success">
                  {notice}
                </p>
              )}
              <Button
                type="submit"
                variant="primary"
                disabled={submitting || code.length !== OTP_LENGTH || email.trim() === ''}
              >
                {submitting ? t('auth.verifying') : t('auth.verifyEmail')}
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={() => void handleResend()}
                disabled={secondsLeft > 0 || email.trim() === ''}
              >
                {secondsLeft > 0
                  ? t('auth.resendIn').replace('{s}', String(secondsLeft))
                  : t('auth.resendCode')}
              </Button>
            </form>
          </>
        )}
      </AuthShell>
    </Container>
  );
}
