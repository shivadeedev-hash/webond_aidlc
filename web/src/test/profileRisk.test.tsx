import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import PersonalInfoCard from '../components/profile/PersonalInfoCard';
import AssetDetail from '../pages/AssetDetail';
import { ApiClientError } from '../lib/errors';
import type { Profile, RiskSummary } from '../types';
import { renderWithProviders } from './renderApp';

const { getProfile, updateProfile, getRisk, submitRisk } = vi.hoisted(() => ({
  getProfile: vi.fn(),
  updateProfile: vi.fn(),
  getRisk: vi.fn(),
  submitRisk: vi.fn(),
}));

vi.mock('../services/profile', () => ({
  profileService: { getProfile, updateProfile, getRisk, submitRisk },
}));

const PROFILE: Profile = {
  id: 42,
  email: 'real.user@webond.test',
  fullName: 'Real User',
  mobile: '0800000000',
  bankAccountMasked: 'xxxxxx7890',
  status: 'active',
};
const VALID_RISK: RiskSummary = {
  riskLevel: 2,
  assessedAt: '2026-10-01T03:00:00Z',
  expiresAt: '2028-10-01T03:00:00Z',
  valid: true,
};

describe('PersonalInfoCard uses the real profile (not the mock user)', () => {
  beforeEach(() => {
    getProfile.mockReset().mockResolvedValue(PROFILE);
    updateProfile.mockReset();
    getRisk.mockReset().mockResolvedValue(VALID_RISK);
    submitRisk.mockReset();
  });

  it("shows the signed-in user's own details and risk level", async () => {
    renderWithProviders(<PersonalInfoCard />);

    expect(await screen.findByDisplayValue('Real User')).toBeInTheDocument();
    expect(screen.getByDisplayValue('real.user@webond.test')).toBeInTheDocument();
    expect(screen.getByText(/Current: xxxxxx7890/)).toBeInTheDocument();
    expect(screen.getByText(/Risk level: 2/)).toBeInTheDocument();
    // The mock demo user is level 4; it must not leak through.
    expect(screen.queryByText(/Risk level: 4/)).not.toBeInTheDocument();
  });

  it('saves name and mobile WITHOUT resending the bank account when it was left blank', async () => {
    updateProfile.mockResolvedValue({ ...PROFILE, fullName: 'New Name' });
    renderWithProviders(<PersonalInfoCard />);

    const name = await screen.findByDisplayValue('Real User');
    fireEvent.change(name, { target: { value: 'New Name' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(updateProfile).toHaveBeenCalledTimes(1));
    expect(updateProfile.mock.calls[0][0]).toEqual({
      fullName: 'New Name',
      mobile: '0800000000',
      bankAccount: undefined,
    });
    expect(await screen.findByRole('status')).toHaveTextContent('Saved');
  });

  it('sends a new bank account when the user types one', async () => {
    updateProfile.mockResolvedValue(PROFILE);
    renderWithProviders(<PersonalInfoCard />);

    await screen.findByDisplayValue('Real User');
    fireEvent.change(screen.getByPlaceholderText('xxxxxx7890'), {
      target: { value: ' 1112223334 ' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(updateProfile).toHaveBeenCalled());
    expect(updateProfile.mock.calls[0][0].bankAccount).toBe('1112223334');
  });

  it('shows the API error when saving fails', async () => {
    updateProfile.mockRejectedValue(
      new ApiClientError('VALIDATION_ERROR', 'Full name is required'),
    );
    renderWithProviders(<PersonalInfoCard />);

    await screen.findByDisplayValue('Real User');
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Full name is required');
  });

  it('tells a user with no assessment to take it, and stores the result via the API', async () => {
    getRisk.mockResolvedValue(null);
    submitRisk.mockResolvedValue({ ...VALID_RISK, riskLevel: 3 });
    renderWithProviders(<PersonalInfoCard />);

    expect(await screen.findByText(/not taken the risk assessment/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Take risk assessment' }));

    // Answer all five questions with the middle option (score 3 → level 3).
    const radios = await screen.findAllByRole('radio');
    expect(radios).toHaveLength(25);
    for (let q = 0; q < 5; q++) fireEvent.click(radios[q * 5 + 2]);
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(submitRisk).toHaveBeenCalledTimes(1));
    const [answers, level] = submitRisk.mock.calls[0];
    expect(level).toBe(3);
    expect(Object.values(answers)).toEqual([3, 3, 3, 3, 3]);
    // Back on the form, now showing the new level.
    expect(await screen.findByText(/Risk level: 3/)).toBeInTheDocument();
  });

  it('flags an expired assessment', async () => {
    getRisk.mockResolvedValue({ ...VALID_RISK, valid: false });
    renderWithProviders(<PersonalInfoCard />);

    expect(await screen.findByText(/assessment has expired/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Retake risk assessment' })).toBeInTheDocument();
  });
});

describe('AssetDetail gates trading on the real assessment', () => {
  beforeEach(() => {
    getProfile.mockReset().mockResolvedValue(PROFILE);
    getRisk.mockReset();
  });

  function renderDetail(): void {
    renderWithProviders(
      <Routes>
        <Route path="/assets/:type/:id" element={<AssetDetail />} />
      </Routes>,
      { initialEntries: ['/assets/fund/3'], authenticated: true },
    );
  }

  it('shows a banner pointing to the profile when the user has no assessment', async () => {
    getRisk.mockResolvedValue(null);
    renderDetail();

    const banner = await screen.findByText(/need a valid risk assessment/i);
    expect(banner).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to Profile' })).toHaveAttribute('href', '/profile');
  });

  it('shows no banner when the assessment is valid', async () => {
    getRisk.mockResolvedValue(VALID_RISK);
    renderDetail();

    await screen.findByRole('heading', { name: /TMBMONEY/ });
    await waitFor(() => expect(getRisk).toHaveBeenCalled());
    expect(screen.queryByText(/need a valid risk assessment/i)).not.toBeInTheDocument();
  });
});
