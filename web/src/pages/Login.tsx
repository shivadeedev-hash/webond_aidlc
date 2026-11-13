// Login (US-003, wireframe 01). Generic invalid-credential message (EX-01),
// lockout notice, links to Register / Forgot password. Uses authService + useAuth.

import { useState, type FormEvent } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import Container from '../components/layout/Container';
import AuthShell from '../components/auth/AuthShell';
import { TextField } from '../components/ui/Field';
import Button from '../components/ui/Button';
import { useAuth } from '../hooks/useAuth';
import { useI18n } from '../i18n/useI18n';
import { ApiClientError } from '../lib/errors';
import { SHOW_DEMO_HINT } from '../lib/constants';
import demoData from '../mocks/user.json';

interface LocationState {
  from?: string;
}

export default function Login(): JSX.Element {
  const { login, isLoading } = useAuth();
  const { t } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [needsVerify, setNeedsVerify] = useState(false);

  const from = (location.state as LocationState | null)?.from ?? '/dashboard';
  const demo = demoData.demoCredentials;

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setError(null);
    setNeedsVerify(false);
    try {
      await login({ email, password });
      navigate(from, { replace: true });
    } catch (err) {
      if (err instanceof ApiClientError && err.code === 'FORBIDDEN') {
        // Right password, but the e-mail is not verified yet (BR-01).
        setNeedsVerify(true);
        setError(t('auth.notVerified'));
      } else if (err instanceof ApiClientError && /locked/i.test(err.message)) {
        // Account locked after repeated failures (EX-01).
        setError('Account locked for 15 minutes after too many failed attempts.');
      } else {
        // Generic message — never reveal which field was wrong (EX-01).
        setError('Invalid email or password');
      }
    }
  }

  return (
    <Container className="py-section">
      <AuthShell>
        <div className="flex items-baseline gap-3">
          <h1 className="text-2xl font-semibold">{t('auth.login')}</h1>
          <span className="text-text-muted">|</span>
          <Link to="/register" className="text-text-muted hover:text-primary hover:underline">
            {t('auth.register')}
          </Link>
        </div>

        <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
          <TextField
            label={t('auth.email')}
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <TextField
            label={t('auth.password')}
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <div className="flex justify-end">
            <Link to="/forgot-password" className="text-sm text-primary hover:underline">
              {t('auth.forgotPassword')}
            </Link>
          </div>

          {error && (
            <p role="alert" className="text-sm text-danger">
              {error}
            </p>
          )}
          {needsVerify && (
            <Link
              to="/verify-email"
              state={{ email: email.trim() }}
              className="text-sm font-medium text-primary hover:underline"
            >
              {t('auth.verifyNow')}
            </Link>
          )}

          <Button type="submit" variant="primary" disabled={isLoading}>
            {isLoading ? t('common.loading') : t('auth.login')}
          </Button>
        </form>

        <hr className="my-6 border-border" />

        <p className="text-sm text-text-muted">
          Register: full name, email, mobile, date of birth, password, accept T&amp;C and PDPA
          consent.
        </p>
        <Link to="/register" className="mt-3 block">
          <Button variant="outline" className="w-full">
            {t('auth.createAccount')}
          </Button>
        </Link>

        {SHOW_DEMO_HINT && (
          <p className="mt-6 rounded-md bg-surface-muted p-3 text-sm text-text-muted">
            Demo: <span className="font-medium text-text">{demo.email}</span> /{' '}
            <span className="font-medium text-text">{demo.password}</span>
          </p>
        )}
      </AuthShell>
    </Container>
  );
}
