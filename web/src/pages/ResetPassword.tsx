// Reset password (US-004, Add-on 2026-11-12b). Takes the e-mail, the 6-digit
// code from the reset e-mail and a new password; on success the user goes back
// to Login. A new code can be requested after 60 s.
import { useState, type FormEvent } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
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
import { RESEND_COOLDOWN_SECONDS } from './VerifyEmail';

export default function ResetPassword(): JSX.Element {
  const { t } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const navState = location.state as { email?: string; sent?: boolean } | null;
  const stateEmail = navState?.email ?? '';

  const [email, setEmail] = useState(stateEmail);
  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [done, setDone] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const { secondsLeft, start } = useCooldown(stateEmail ? RESEND_COOLDOWN_SECONDS : 0);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);
    setNotice(null);
    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }
    if (newPassword !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    setSubmitting(true);
    try {
      await authService.resetPassword({ email: email.trim(), code, newPassword });
      setDone(true);
      setTimeout(() => navigate('/login', { replace: true }), 1500);
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleResend(): Promise<void> {
    setError(null);
    try {
      await authService.forgotPassword({ email: email.trim() });
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
        <h1 className="text-2xl font-bold">{t('auth.resetPassword')}</h1>
        {done ? (
          <p
            className="mt-6 inline-flex items-center gap-2 text-sm font-medium text-success"
            role="status"
          >
            <CheckCircle size={18} />
            {t('auth.resetSuccess')}
          </p>
        ) : (
          <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
            {navState?.sent && <p className="text-sm text-text-muted">{t('auth.forgotSent')}</p>}
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
            <TextField
              label={t('auth.newPassword')}
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
              required
            />
            <TextField
              label={t('auth.confirmPassword')}
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              autoComplete="new-password"
              required
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
              {submitting ? t('common.loading') : t('auth.resetPassword')}
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
            <Link to="/login" className="text-center text-sm text-primary hover:underline">
              {t('auth.login')}
            </Link>
          </form>
        )}
      </AuthShell>
    </Container>
  );
}
