// Forgot password (US-004). Always reports success (no account enumeration);
// a 6-digit reset code is e-mailed, then the user enters it on Reset password.

import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import Container from '../components/layout/Container';
import AuthShell from '../components/auth/AuthShell';
import { TextField } from '../components/ui/Field';
import Button from '../components/ui/Button';
import { authService } from '../services/auth';
import { useI18n } from '../i18n/useI18n';
import { ApiClientError } from '../lib/errors';
import { CheckCircle } from '../components/ui/icons';

export default function ForgotPassword(): JSX.Element {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await authService.forgotPassword({ email: email.trim() });
      // Straight to the code form with the e-mail filled in.
      navigate('/reset-password', { state: { email: email.trim(), sent: true } });
      setSent(true);
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Container className="py-section">
      <AuthShell>
        <h1 className="text-2xl font-semibold">{t('auth.resetPassword')}</h1>

        {sent ? (
          <div className="mt-6">
            <p className="inline-flex items-center gap-2 text-sm font-medium text-success">
              <CheckCircle size={18} />
              {t('auth.forgotSent')}
            </p>
            <Link to="/reset-password" state={{ email: email.trim() }} className="mt-4 block">
              <Button variant="primary" className="w-full">
                {t('auth.enterCode')}
              </Button>
            </Link>
            <Link
              to="/login"
              className="mt-2 block text-center text-sm text-primary hover:underline"
            >
              {t('auth.login')}
            </Link>
          </div>
        ) : (
          <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
            <TextField
              label={t('auth.email')}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
            {error && (
              <p role="alert" className="text-sm text-danger">
                {error}
              </p>
            )}
            <Button type="submit" variant="primary" disabled={submitting}>
              {submitting ? t('common.loading') : t('auth.sendResetLink')}
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
