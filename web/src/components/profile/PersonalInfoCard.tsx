// Personal information + risk suitability card (US-006, US-007). Everything comes
// from the profile service, so in API mode it shows the signed-in user's real
// data: an account that never took the questionnaire says so instead of
// borrowing someone else's level.

import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { profileService } from '../../services/profile';
import { ApiClientError } from '../../lib/errors';
import { formatDateBangkok } from '../../lib/format';
import { useI18n } from '../../i18n/useI18n';
import type { Profile, RiskLevel, RiskSummary } from '../../types';
import { Card, CardBody, CardHeader, CardTitle } from '../ui/Card';
import Button from '../ui/Button';
import { TextField } from '../ui/Field';
import { ErrorState, LoadingState } from '../ui/StateViews';
import RiskQuestionnaire from '../risk/RiskQuestionnaire';

const RISK_LABELS: Record<RiskLevel, string> = {
  1: 'Low',
  2: 'Moderately low',
  3: 'Moderate',
  4: 'Moderately high',
  5: 'High',
};

function messageOf(err: unknown, fallback: string): string {
  return err instanceof ApiClientError ? err.message : fallback;
}

export default function PersonalInfoCard(): JSX.Element {
  const { t } = useI18n();
  const profileQuery = useQuery({
    queryKey: ['profile'],
    queryFn: () => profileService.getProfile(),
  });
  const riskQuery = useQuery({ queryKey: ['risk'], queryFn: () => profileService.getRisk() });

  let body: JSX.Element;
  if (profileQuery.isLoading || riskQuery.isLoading) {
    body = <LoadingState />;
  } else if (profileQuery.isError || riskQuery.isError || !profileQuery.data) {
    body = (
      <ErrorState
        onRetry={() => {
          void profileQuery.refetch();
          void riskQuery.refetch();
        }}
      />
    );
  } else {
    body = (
      <ProfileForm
        key={profileQuery.data.id}
        profile={profileQuery.data}
        risk={riskQuery.data ?? null}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('page.profile.personalInfo')}</CardTitle>
      </CardHeader>
      <CardBody>{body}</CardBody>
    </Card>
  );
}

function ProfileForm({
  profile,
  risk,
}: {
  profile: Profile;
  risk: RiskSummary | null;
}): JSX.Element {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [fullName, setFullName] = useState(profile.fullName);
  const [mobile, setMobile] = useState(profile.mobile);
  const [bankAccount, setBankAccount] = useState('');
  const [retaking, setRetaking] = useState(false);

  const save = useMutation({
    mutationFn: () =>
      profileService.updateProfile({
        fullName,
        mobile,
        // Blank = keep the account on file (the field never shows the full number).
        bankAccount: bankAccount.trim() === '' ? undefined : bankAccount.trim(),
      }),
    onSuccess: (updated) => {
      queryClient.setQueryData(['profile'], updated);
      setBankAccount('');
    },
  });

  const assess = useMutation({
    mutationFn: (v: { level: RiskLevel; answers: Record<string, number> }) =>
      profileService.submitRisk(v.answers, v.level),
    onSuccess: (latest) => {
      queryClient.setQueryData(['risk'], latest);
      setRetaking(false);
    },
  });

  if (retaking) {
    return (
      <div className="flex flex-col gap-3">
        <RiskQuestionnaire
          onCancel={() => setRetaking(false)}
          onComplete={(level, answers) => assess.mutate({ level, answers })}
        />
        {assess.isError && (
          <p role="alert" className="text-sm text-danger">
            {messageOf(assess.error, t('common.error'))}
          </p>
        )}
      </div>
    );
  }

  function handleSubmit(e: FormEvent): void {
    e.preventDefault();
    save.mutate();
  }

  const missing = risk === null;
  const expired = risk !== null && !risk.valid;
  const needsAttention = missing || expired;

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <TextField
          label={t('auth.fullName')}
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          autoComplete="name"
        />
        <TextField
          label={t('auth.mobile')}
          value={mobile}
          onChange={(e) => setMobile(e.target.value)}
          inputMode="tel"
        />
      </div>
      <TextField
        label={`${t('page.profile.emailVerified')} — OK`}
        value={profile.email}
        readOnly
        disabled
      />
      <TextField
        label={t('page.profile.bankAccount')}
        value={bankAccount}
        onChange={(e) => setBankAccount(e.target.value)}
        placeholder={profile.bankAccountMasked}
        hint={profile.bankAccountMasked ? `Current: ${profile.bankAccountMasked}` : undefined}
      />

      <div
        className={
          needsAttention
            ? 'rounded-md border border-warning bg-warning/10 p-3 text-sm'
            : 'rounded-md border border-success bg-success/10 p-3 text-sm'
        }
      >
        {missing && <p className="font-medium text-text">{t('page.profile.noRisk')}</p>}
        {expired && risk && (
          <p className="font-medium text-text">
            {t('page.profile.riskExpired')} ({t('page.profile.expires')}{' '}
            {formatDateBangkok(risk.expiresAt)})
          </p>
        )}
        {risk && !expired && (
          <>
            <p className="font-medium text-success">
              {t('page.profile.riskLevel')}: {risk.riskLevel} – {RISK_LABELS[risk.riskLevel]}
            </p>
            <p className="mt-0.5 text-text-muted">
              {t('page.profile.assessed')}: {formatDateBangkok(risk.assessedAt)} (
              {t('page.profile.expires')} {formatDateBangkok(risk.expiresAt)})
            </p>
          </>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Button
          variant={needsAttention ? 'primary' : 'outline'}
          type="button"
          onClick={() => setRetaking(true)}
        >
          {missing ? t('page.profile.takeRisk') : t('page.profile.retakeRisk')}
        </Button>
        <Button
          variant={needsAttention ? 'outline' : 'primary'}
          type="submit"
          disabled={save.isPending}
        >
          {save.isPending ? t('common.loading') : t('common.save')}
        </Button>
        {save.isSuccess && (
          <span role="status" className="text-sm text-success">
            {t('page.profile.saved')}
          </span>
        )}
      </div>
      {save.isError && (
        <p role="alert" className="text-sm text-danger">
          {messageOf(save.error, t('common.error'))}
        </p>
      )}
    </form>
  );
}
