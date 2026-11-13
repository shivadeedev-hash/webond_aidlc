// Register (US-001, wireframe 01). Collects email, password, full name, mobile,
// date of birth + required PDPA consent. Duplicate-email → EX-02 message.
// On success, routes to the email-verification step (US-002).

import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import Container from '../components/layout/Container';
import AuthShell from '../components/auth/AuthShell';
import { TextField } from '../components/ui/Field';
import Button from '../components/ui/Button';
import { authService } from '../services/auth';
import { useI18n } from '../i18n/useI18n';
import { ApiClientError } from '../lib/errors';

interface Errors {
  email?: string;
  password?: string;
  confirm?: string;
  fullName?: string;
  mobile?: string;
  dateOfBirth?: string;
  consent?: string;
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function Register(): JSX.Element {
  const { t } = useI18n();
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [fullName, setFullName] = useState('');
  const [mobile, setMobile] = useState('');
  const [dateOfBirth, setDateOfBirth] = useState('');
  const [consent, setConsent] = useState(false);

  const [errors, setErrors] = useState<Errors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function validate(): Errors {
    const next: Errors = {};
    if (!EMAIL_RE.test(email)) next.email = 'Enter a valid email address.';
    if (password.length < 8) next.password = 'Password must be at least 8 characters.';
    if (confirm !== password) next.confirm = 'Passwords do not match.';
    if (fullName.trim().length < 2) next.fullName = 'Enter your full name.';
    if (!/^0\d{8,9}$/.test(mobile)) next.mobile = 'Enter a valid mobile number.';
    if (!dateOfBirth) next.dateOfBirth = 'Select your date of birth.';
    if (!consent) next.consent = 'You must accept the T&C and PDPA consent to continue.';
    return next;
  }

  async function handleSubmit(e: FormEvent): Promise<void> {
    e.preventDefault();
    setFormError(null);
    const found = validate();
    setErrors(found);
    if (Object.keys(found).length > 0) return;

    setSubmitting(true);
    try {
      await authService.register({ email, password, fullName, mobile, dateOfBirth, consent });
      navigate('/verify-email', { state: { email } });
    } catch (err) {
      if (err instanceof ApiClientError && err.code === 'DUPLICATE_EMAIL') {
        setErrors((prev) => ({ ...prev, email: 'This email is already registered' }));
      } else {
        setFormError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Container className="py-section">
      <AuthShell>
        <div className="flex items-baseline gap-3">
          <Link to="/login" className="text-text-muted hover:text-primary hover:underline">
            {t('auth.login')}
          </Link>
          <span className="text-text-muted">|</span>
          <h1 className="text-2xl font-semibold">{t('auth.register')}</h1>
        </div>

        <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
          <TextField
            label={t('auth.fullName')}
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            error={errors.fullName}
            autoComplete="name"
          />
          <TextField
            label={t('auth.email')}
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
            autoComplete="email"
          />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextField
              label={t('auth.mobile')}
              value={mobile}
              onChange={(e) => setMobile(e.target.value)}
              error={errors.mobile}
              autoComplete="tel"
              inputMode="tel"
            />
            <TextField
              label={t('auth.dateOfBirth')}
              type="date"
              value={dateOfBirth}
              onChange={(e) => setDateOfBirth(e.target.value)}
              error={errors.dateOfBirth}
            />
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextField
              label={t('auth.password')}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              error={errors.password}
              autoComplete="new-password"
            />
            <TextField
              label={t('auth.confirmPassword')}
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              error={errors.confirm}
              autoComplete="new-password"
            />
          </div>

          <div>
            <label className="flex items-start gap-2 text-sm">
              <input
                type="checkbox"
                checked={consent}
                onChange={(e) => setConsent(e.target.checked)}
                className="mt-0.5 h-5 w-5 rounded border-border accent-[color:var(--color-primary)]"
                aria-invalid={errors.consent ? true : undefined}
              />
              <span>{t('auth.consent')}</span>
            </label>
            {errors.consent && (
              <p role="alert" className="mt-1 text-xs text-danger">
                {errors.consent}
              </p>
            )}
          </div>

          {formError && (
            <p role="alert" className="text-sm text-danger">
              {formError}
            </p>
          )}

          <Button type="submit" variant="primary" disabled={submitting}>
            {submitting ? t('common.loading') : t('auth.createAccount')}
          </Button>
        </form>

        <p className="mt-6 text-sm text-text-muted">
          {t('auth.haveAccount')}{' '}
          <Link to="/login" className="text-primary hover:underline">
            {t('auth.login')}
          </Link>
        </p>
      </AuthShell>
    </Container>
  );
}
