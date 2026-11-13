// Profile & Wallet (US-006..010, wireframe 09).
// Left: personal info (name/mobile/verified email), masked withdrawal bank
// account, risk level + retake assessment. Right: wallet balances
// (balance/reserved/available), deposit/withdraw, and wallet history.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { walletService } from '../services/wallet';
import { useI18n } from '../i18n/useI18n';
import { formatTHB } from '../lib/format';
import PageHeader from '../components/ui/PageHeader';
import { Card, CardBody, CardHeader, CardTitle } from '../components/ui/Card';
import Button from '../components/ui/Button';
import { ErrorState } from '../components/ui/StateViews';
import PersonalInfoCard from '../components/profile/PersonalInfoCard';
import WalletHistory from '../components/wallet/WalletHistory';
import WalletActionDialog from '../components/wallet/WalletActionDialog';

export default function Profile(): JSX.Element {
  const { t } = useI18n();

  const [dialog, setDialog] = useState<'deposit' | 'withdraw' | null>(null);

  const walletQuery = useQuery({
    queryKey: ['wallet'],
    queryFn: () => walletService.getWallet(),
  });
  const txQuery = useQuery({
    queryKey: ['walletTx'],
    queryFn: () => walletService.listTransactions(1, 20),
  });

  const wallet = walletQuery.data;
  const available = Number(wallet?.available ?? 0);

  function refreshWallet(): void {
    setDialog(null);
    void walletQuery.refetch();
    void txQuery.refetch();
  }

  if (walletQuery.isError) {
    return <ErrorState onRetry={() => void walletQuery.refetch()} />;
  }

  return (
    <div>
      <PageHeader title={t('page.profile.title')} />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Personal information + risk */}
        <PersonalInfoCard />

        {/* Wallet */}
        <Card>
          <CardHeader>
            <CardTitle>{t('page.profile.wallet')}</CardTitle>
          </CardHeader>
          <CardBody>
            <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm">
              <div>
                <dt className="text-text-muted">{t('page.profile.balance')}</dt>
                <dd className="font-semibold">{formatTHB(wallet?.balance ?? 0)}</dd>
              </div>
              <div>
                <dt className="text-text-muted">{t('page.profile.reserved')}</dt>
                <dd className="font-semibold">{formatTHB(wallet?.reserved ?? 0)}</dd>
              </div>
              <div>
                <dt className="text-text-muted">{t('page.profile.available')}</dt>
                <dd className="font-semibold text-primary">{formatTHB(available)}</dd>
              </div>
            </dl>

            <div className="mt-4 grid grid-cols-2 gap-3">
              <Button variant="primary" onClick={() => setDialog('deposit')}>
                {t('page.profile.deposit')}
              </Button>
              <Button variant="outline" onClick={() => setDialog('withdraw')}>
                {t('page.profile.withdraw')}
              </Button>
            </div>

            <div className="mt-5">
              <WalletHistory transactions={txQuery.data?.items ?? []} loading={txQuery.isLoading} />
            </div>
          </CardBody>
        </Card>
      </div>

      {dialog && (
        <WalletActionDialog
          mode={dialog}
          available={available}
          onClose={() => setDialog(null)}
          onSuccess={refreshWallet}
        />
      )}
    </div>
  );
}
