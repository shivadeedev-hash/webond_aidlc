import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { I18nProvider } from '../i18n/I18nProvider';
import VerifyEmail from '../pages/VerifyEmail';
import ResetPassword from '../pages/ResetPassword';
import { ApiClientError } from '../lib/errors';

const { verify, resendVerification, resetPassword, forgotPassword } = vi.hoisted(() => ({
  verify: vi.fn(),
  resendVerification: vi.fn(),
  resetPassword: vi.fn(),
  forgotPassword: vi.fn(),
}));
vi.mock('../services/auth', () => ({
  authService: { verify, resendVerification, resetPassword, forgotPassword },
}));

function renderAt(path: string, state?: unknown): void {
  render(
    <I18nProvider initialLocale="en">
      <MemoryRouter initialEntries={[{ pathname: path, state }]}>
        <Routes>
          <Route path="/verify-email" element={<VerifyEmail />} />
          <Route path="/reset-password" element={<ResetPassword />} />
          <Route path="/login" element={<p>login page</p>} />
        </Routes>
      </MemoryRouter>
    </I18nProvider>,
  );
}

describe('VerifyEmail with a 6-digit OTP (US-002, Add-on 2026-11-12b)', () => {
  beforeEach(() => {
    verify.mockReset();
    resendVerification.mockReset();
  });

  it('prefills the e-mail, keeps only digits, and verifies with email + code', async () => {
    verify.mockResolvedValue(undefined);
    renderAt('/verify-email', { email: 'new@example.com' });
    expect(screen.getByLabelText('Email')).toHaveValue('new@example.com');

    const submit = screen.getByRole('button', { name: 'Verify email' });
    expect(submit).toBeDisabled();
    fireEvent.change(screen.getByLabelText('6-digit code'), { target: { value: '12 34-56x9' } });
    expect(screen.getByLabelText('6-digit code')).toHaveValue('123456');
    fireEvent.click(submit);

    expect(
      await screen.findByText('Email verified. Your account is now active.'),
    ).toBeInTheDocument();
    expect(verify).toHaveBeenCalledWith({ email: 'new@example.com', code: '123456' });
  });

  it('shows the API message for a wrong code', async () => {
    verify.mockImplementation(() =>
      Promise.reject(
        new ApiClientError('VALIDATION_ERROR', 'Incorrect code. 4 attempt(s) left.', {
          status: 422,
        }),
      ),
    );
    renderAt('/verify-email', { email: 'new@example.com' });
    fireEvent.change(screen.getByLabelText('6-digit code'), { target: { value: '000000' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verify email' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Incorrect code. 4 attempt(s) left.',
    );
  });

  it('resend waits for the 60 s cooldown after registration, then sends a new code', async () => {
    vi.useFakeTimers();
    try {
      resendVerification.mockResolvedValue(undefined);
      renderAt('/verify-email', { email: 'new@example.com' });
      expect(screen.getByRole('button', { name: /Send a new code in 60s/ })).toBeDisabled();
      for (let i = 0; i < 60; i++) {
        await act(async () => {
          vi.advanceTimersByTime(1000);
        });
      }
      const resend = screen.getByRole('button', { name: 'Send a new code' });
      expect(resend).toBeEnabled();
      await act(async () => {
        fireEvent.click(resend);
      });
      expect(resendVerification).toHaveBeenCalledWith({ email: 'new@example.com' });
      expect(screen.getByRole('status')).toHaveTextContent('A new code is on its way');
    } finally {
      vi.useRealTimers();
    }
  });
});

describe('ResetPassword with a 6-digit OTP (US-004)', () => {
  beforeEach(() => resetPassword.mockReset() && undefined);

  it('sends email + code + new password', async () => {
    resetPassword.mockResolvedValue(undefined);
    renderAt('/reset-password', { email: 'me@example.com', sent: true });
    expect(screen.getByText(/we sent a 6-digit code/i)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('6-digit code'), { target: { value: '654321' } });
    fireEvent.change(screen.getByLabelText('New password'), { target: { value: 'NewPass@123' } });
    fireEvent.change(screen.getByLabelText('Confirm password'), {
      target: { value: 'NewPass@123' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Reset password' }));
    expect(await screen.findByText(/Password reset successful/)).toBeInTheDocument();
    expect(resetPassword).toHaveBeenCalledWith({
      email: 'me@example.com',
      code: '654321',
      newPassword: 'NewPass@123',
    });
  });
});
