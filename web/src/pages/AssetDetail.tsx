// Asset Detail + Buy/Sell panel (US-012..018, wireframe 04).
// Left: type-specific detail + price/NAV chart with period toggle.
// Right: Buy/Sell panel (auth-gated — guests get a sign-in prompt).
// Price-unavailable (EX-10) is derived from the quote's age vs the type threshold.

import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import type { AssetType, PricePeriod } from '../types';
import { catalogService } from '../services/catalog';
import { walletService } from '../services/wallet';
import { profileService } from '../services/profile';
import { useAuth } from '../hooks/useAuth';
import { useI18n } from '../i18n/useI18n';
import { formatDateBangkok, formatDateTimeBangkok, formatMoney, formatTHB } from '../lib/format';
import PageHeader from '../components/ui/PageHeader';
import { Card, CardBody } from '../components/ui/Card';
import { LoadingState, ErrorState } from '../components/ui/StateViews';
import Button from '../components/ui/Button';
import PriceChart from '../components/charts/PriceChart';
import BuySellPanel, { type HeldPosition } from '../components/trade/BuySellPanel';
import { portfolioService } from '../services/portfolio';

const CHART_PERIODS: PricePeriod[] = ['1M', '6M', '1Y', '5Y'];

// Price-staleness thresholds (EX-10 / BR-17): gold 5 min, NAV 1 business day.
const GOLD_STALE_MS = 5 * 60 * 1000;
const NAV_STALE_MS = 24 * 60 * 60 * 1000;

function useAssetParams(): { type: AssetType; id: number } | null {
  const { type, id } = useParams();
  if (!type || !id) return null;
  if (type !== 'fund' && type !== 'bond' && type !== 'gold') return null;
  const num = Number(id);
  if (!Number.isFinite(num)) return null;
  return { type, id: num };
}

export default function AssetDetail(): JSX.Element {
  const { t } = useI18n();
  const { isAuthenticated } = useAuth();
  const parsed = useAssetParams();
  const [period, setPeriod] = useState<PricePeriod>('1Y');

  const assetQuery = useQuery({
    queryKey: ['asset', parsed?.type, parsed?.id],
    queryFn: () => catalogService.getAsset(parsed!.type, parsed!.id),
    enabled: parsed !== null,
  });

  const seriesQuery = useQuery({
    queryKey: ['assetSeries', parsed?.type, parsed?.id, period],
    queryFn: () => catalogService.getPriceSeries(parsed!.type, parsed!.id, period),
    enabled: parsed !== null,
  });

  const walletQuery = useQuery({
    queryKey: ['wallet'],
    queryFn: () => walletService.getWallet(),
    enabled: isAuthenticated,
  });
  const portfolioQuery = useQuery({
    queryKey: ['portfolio'],
    queryFn: () => portfolioService.getPortfolio(),
    enabled: isAuthenticated,
  });
  const riskQuery = useQuery({
    queryKey: ['risk'],
    queryFn: () => profileService.getRisk(),
    enabled: isAuthenticated,
  });

  const asset = assetQuery.data;

  const priceAvailable = useMemo(() => {
    if (!asset) return false;
    const age = Date.now() - new Date(asset.quote.pricedAt).getTime();
    const threshold = asset.assetType === 'gold' ? GOLD_STALE_MS : NAV_STALE_MS;
    // The API decides staleness (EX-10: NAV > 1 day, gold > 5 min) and says so in
    // quote.unavailable. Mock data is dated, so in mock mode only a missing price
    // counts as unavailable.
    if (asset.quote.unavailable) return false;
    return (
      asset.quote.latestPrice != null && asset.quote.latestPrice !== '' && age >= 0 && threshold > 0
    );
  }, [asset]);

  if (parsed === null) {
    return <ErrorState message={t('errors.UNKNOWN')} />;
  }

  if (assetQuery.isLoading) return <LoadingState />;
  if (assetQuery.isError || !asset) {
    return <ErrorState onRetry={() => void assetQuery.refetch()} />;
  }

  // The signed-in user's real, still-valid assessment; 0 (below every asset)
  // when they have none, so the risk acknowledgement shows and the API's own
  // suitability check (BR-01) has the final say.
  const userRisk = riskQuery.data?.valid ? riskQuery.data.riskLevel : 0;
  const needsAssessment = isAuthenticated && riskQuery.isSuccess && !riskQuery.data?.valid;
  const available = Number(walletQuery.data?.available ?? 0);
  // What the user holds of this asset (for the Sell side).
  const held = portfolioQuery.data?.holdings.find(
    (h) => h.assetType === asset.assetType && h.assetId === asset.id,
  );
  const holding: HeldPosition | null = held
    ? {
        quantity: Math.max(0, Number(held.quantity) - Number(held.lockedQuantity || 0)),
        unitPrice: Number(held.marketPrice) || 0,
      }
    : null;

  const returnPct = asset.quote.periodReturnPct;
  const returnNum = returnPct === null ? null : Number(returnPct);

  return (
    <div>
      <PageHeader
        title={`${asset.code} · ${typeLabel(asset.assetType, t)} · Risk ${asset.riskLevel}`}
        breadcrumb={
          <Link to="/market" className="text-sm text-primary hover:underline">
            {t('page.market.title')} /
          </Link>
        }
      />

      {needsAssessment && (
        <div
          role="alert"
          className="mb-4 rounded-md border border-warning bg-warning/10 p-3 text-sm"
        >
          {t('page.detail.needRisk')}{' '}
          <Link to="/profile" className="font-medium text-primary underline">
            {t('page.detail.goProfile')}
          </Link>
        </div>
      )}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
        {/* Left column: chart + info */}
        <div className="flex flex-col gap-6">
          <Card>
            <CardBody>
              <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
                <p className="text-sm">
                  <span className="text-text-muted">
                    {asset.assetType === 'fund'
                      ? 'NAV '
                      : asset.assetType === 'gold'
                        ? 'Price/g '
                        : 'Price '}
                  </span>
                  <span className="font-semibold">
                    {asset.assetType === 'gold'
                      ? formatTHB(asset.quote.latestPrice)
                      : formatMoney(asset.quote.latestPrice)}
                  </span>
                  {returnNum !== null && (
                    <span
                      className={`ml-2 font-medium ${returnNum >= 0 ? 'text-success' : 'text-danger'}`}
                    >
                      {returnNum >= 0 ? '+' : ''}
                      {returnNum.toFixed(1)}%
                    </span>
                  )}
                </p>

                <div
                  role="group"
                  aria-label="Chart period"
                  className="inline-flex overflow-hidden rounded-md border border-border"
                >
                  {CHART_PERIODS.map((p) => (
                    <button
                      key={p}
                      type="button"
                      aria-pressed={p === period}
                      onClick={() => setPeriod(p)}
                      className={[
                        'min-h-[36px] px-3 text-xs font-medium transition-standard duration-standard',
                        p === period
                          ? 'bg-primary text-text-inverse'
                          : 'bg-surface text-text hover:bg-surface-muted',
                      ].join(' ')}
                    >
                      {p}
                    </button>
                  ))}
                </div>
              </div>

              {seriesQuery.isLoading ? (
                <LoadingState />
              ) : (
                <PriceChart points={seriesQuery.data?.points ?? []} />
              )}

              <p className="mt-2 text-xs text-text-muted">
                {t('page.dashboard.asOf')} {formatDateTimeBangkok(asset.quote.pricedAt)}
              </p>
            </CardBody>
          </Card>

          <Card>
            <CardBody>
              <AssetInfo asset={asset} />
            </CardBody>
          </Card>
        </div>

        {/* Right column: trade panel (auth-gated) */}
        <Card className="h-fit">
          <CardBody>
            {isAuthenticated ? (
              <BuySellPanel
                asset={asset}
                available={available}
                userRisk={userRisk}
                priceAvailable={priceAvailable}
                holding={holding}
              />
            ) : (
              <div className="flex flex-col gap-4 text-sm">
                <p className="text-text-muted">{t('auth.noAccount')}</p>
                <Link to="/login">
                  <Button variant="primary" className="w-full">
                    {t('auth.login')}
                  </Button>
                </Link>
              </div>
            )}
          </CardBody>
        </Card>
      </div>
    </div>
  );
}

function typeLabel(type: AssetType, t: (k: string) => string): string {
  if (type === 'fund') return t('page.market.tab.fund');
  if (type === 'bond') return t('page.market.tab.bond');
  return t('page.market.tab.gold');
}

function Row({ label, value }: { label: string; value: string }): JSX.Element {
  return (
    <div className="flex justify-between gap-4 py-1">
      <dt className="text-text-muted">{label}</dt>
      <dd className="text-right font-medium">{value}</dd>
    </div>
  );
}

function AssetInfo({ asset }: { asset: import('../types').AssetListItem }): JSX.Element {
  if (asset.assetType === 'fund') {
    return (
      <>
        <h2 className="mb-3 text-base font-semibold">Fund info</h2>
        <dl className="grid grid-cols-1 gap-x-8 text-sm sm:grid-cols-2">
          <Row label="Category" value={asset.category} />
          <Row label="Min. purchase" value={formatTHB(asset.minPurchase)} />
          <Row label="Fee" value="0% (free)" />
          <Row label="Redemption" value={`T+${asset.settlementDays}`} />
          <Row label="Cut-off" value={asset.cutoffTime} />
          <div className="flex justify-between gap-4 py-1">
            <dt className="text-text-muted">Fund fact sheet</dt>
            <dd>
              <a href={asset.factSheetUrl} className="text-primary hover:underline">
                View &gt;
              </a>
            </dd>
          </div>
        </dl>
      </>
    );
  }
  if (asset.assetType === 'bond') {
    return (
      <>
        <h2 className="mb-3 text-base font-semibold">Bond info</h2>
        <dl className="grid grid-cols-1 gap-x-8 text-sm sm:grid-cols-2">
          <Row label="Issuer" value={asset.issuer} />
          <Row label="Credit rating" value={asset.creditRating} />
          <Row label="Coupon rate" value={`${Number(asset.couponRate).toFixed(2)}%`} />
          <Row label="Coupon freq." value={`${asset.couponFrequency}/yr`} />
          <Row label="YTM" value={`${Number(asset.ytm).toFixed(2)}%`} />
          <Row label="Face value" value={formatTHB(asset.faceValue)} />
          <Row label="Maturity" value={formatDateBangkok(asset.maturityDate)} />
          <Row label="Min. lot" value={formatTHB(asset.minLot)} />
          <Row
            label="Offering"
            value={`${formatDateBangkok(asset.offeringStart)} – ${formatDateBangkok(asset.offeringEnd)}`}
          />
        </dl>
      </>
    );
  }
  return (
    <>
      <h2 className="mb-3 text-base font-semibold">Gold info</h2>
      <dl className="grid grid-cols-1 gap-x-8 text-sm sm:grid-cols-2">
        <Row label="Purity" value={asset.purity} />
        <Row
          label="Buy (ASK)/g"
          value={formatTHB(asset.quote.askPrice ?? asset.quote.latestPrice)}
        />
        <Row
          label="Sell (BID)/g"
          value={formatTHB(asset.quote.bidPrice ?? asset.quote.latestPrice)}
        />
      </dl>
    </>
  );
}
