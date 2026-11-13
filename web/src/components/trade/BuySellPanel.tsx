// Buy/Sell panel for Asset Detail (US-013..018, wireframe 04; Add-on 2026-11-12d).
// - Friendly amount field: "1,000" / "฿1000" accepted, live grouping, no minus.
// - Buy in THB. Sell in THB or units (gold: grams), with the holding shown.
// - Quick picks: buy = minimum / common amounts / all cash (bond = lot multiples);
//   sell = 25% / 50% / all.
// - Specific inline messages (minimum, lot multiple, cash, holding) before submit.
// - Fund/Bond → placeOrder; Gold → 30 s quote (QuoteDialog) → confirm.
// - Risk-mismatch ack when asset risk > user risk (EX-12).

import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import type { AssetListItem, GoldQuote, OrderSide } from '../../types';
import { orderService } from '../../services/order';
import { ApiClientError } from '../../lib/errors';
import { formatTHB, formatUnits } from '../../lib/format';
import { amountValue, floorTo } from '../../lib/amountInput';
import { useI18n } from '../../i18n/useI18n';
import Button from '../ui/Button';
import { CheckCircle } from '../ui/icons';
import RiskAckBox from './RiskAckBox';
import QuoteDialog from './QuoteDialog';
import AmountInput, { type QuickPick } from './AmountInput';

/** What the user holds of this asset (units for fund/bond, grams for gold). */
export interface HeldPosition {
  quantity: number;
  /** Current price per unit/gram in THB. */
  unitPrice: number;
}

interface Props {
  asset: AssetListItem;
  /** Available wallet cash (THB). */
  available: number;
  /** The investor's current risk level (1–5). */
  userRisk: number;
  /** Whether a price is available (EX-10). Disables trading when false. */
  priceAvailable: boolean;
  /** Current holding of this asset, or null when none. */
  holding?: HeldPosition | null;
}

type SellMode = 'amount' | 'qty';

const COMMON_BUY = [1000, 5000, 10000];

export default function BuySellPanel({
  asset,
  available,
  userRisk,
  priceAvailable,
  holding = null,
}: Props): JSX.Element {
  const { t } = useI18n();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [side, setSide] = useState<OrderSide>('buy');
  const [sellMode, setSellMode] = useState<SellMode>('amount');
  const [value, setValue] = useState('');
  const [touched, setTouched] = useState(false);
  const [riskAck, setRiskAck] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [quote, setQuote] = useState<GoldQuote | null>(null);
  const [quoteError, setQuoteError] = useState<string | null>(null);

  const isGold = asset.assetType === 'gold';
  const isBond = asset.assetType === 'bond';
  const isFund = asset.assetType === 'fund';
  const riskMismatch = asset.riskLevel > userRisk;
  const byQty = side === 'sell' && sellMode === 'qty';
  const qtyUnit = isGold ? t('trade.gramsSuffix') : t('trade.unitsSuffix');

  const unitPrice = useMemo(() => {
    if (isGold) {
      return Number(
        side === 'buy'
          ? (asset.quote.askPrice ?? asset.quote.latestPrice)
          : (asset.quote.bidPrice ?? asset.quote.latestPrice),
      );
    }
    if (side === 'sell' && holding && holding.unitPrice > 0) return holding.unitPrice;
    return Number(asset.quote.latestPrice);
  }, [asset, isGold, side, holding]);

  const num = amountValue(value);
  const heldQty = holding?.quantity ?? 0;
  const heldValue = heldQty * unitPrice;
  const minBuy = isFund ? Number(asset.minPurchase) || 0 : 0;
  const lot = isBond ? Number(asset.minLot) || 0 : 0;

  // Quantity and THB of the order, whichever way it was entered.
  const orderQty = byQty ? num : unitPrice > 0 ? num / unitPrice : 0;
  const orderThb = byQty ? num * unitPrice : num;

  /** Specific message for the current input, or null when it is fine. */
  const problem = useMemo((): string | null => {
    if (!(num > 0)) return t('trade.errEmpty');
    if (side === 'buy') {
      if (isFund && num < minBuy) return t('trade.errMin').replace('{min}', formatTHB(minBuy));
      if (isBond && lot > 0 && (num < lot || Math.round(num * 100) % Math.round(lot * 100) !== 0)) {
        return t('trade.errLot')
          .replace(/\{lot\}/g, formatTHB(lot))
          .replace('{lot2}', formatTHB(lot * 2));
      }
      if (num > available) return t('trade.errCash').replace('{cash}', formatTHB(available));
      return null;
    }
    if (heldQty <= 0) return t('trade.errNoHolding');
    const over = byQty ? num > heldQty + 1e-9 : num > heldValue + 0.005;
    if (over) {
      return t('trade.errOverHolding').replace(
        '{max}',
        `${formatUnits(heldQty)} ${qtyUnit} (≈ ${formatTHB(heldValue)})`,
      );
    }
    return null;
  }, [num, side, isFund, isBond, minBuy, lot, available, heldQty, heldValue, byQty, qtyUnit, t]);

  const picks = useMemo((): QuickPick[] => {
    if (side === 'sell') {
      if (heldQty <= 0) return [];
      return [
        { label: '25%', p: 0.25 },
        { label: '50%', p: 0.5 },
        { label: t('trade.all'), p: 1 },
      ].map(({ label, p }) => ({
        label,
        value: byQty ? floorTo(heldQty * p, 4) : floorTo(heldValue * p, 2),
      }));
    }
    if (isBond) {
      return lot > 0
        ? [1, 2, 5]
            .filter((k) => lot * k <= available)
            .map((k) => ({ label: formatTHB(lot * k), value: String(lot * k) }))
        : [];
    }
    const base =
      isFund && minBuy > 0 ? [minBuy, ...COMMON_BUY.filter((v) => v > minBuy)] : COMMON_BUY;
    const list = base
      .filter((v) => v <= available)
      .map((v) => ({ label: formatTHB(v), value: String(v) }));
    if (available >= Math.max(minBuy, 1))
      list.push({ label: t('trade.all'), value: floorTo(available, 2) });
    return list;
  }, [side, heldQty, heldValue, byQty, isBond, isFund, lot, minBuy, available, t]);

  const showProblem = touched && problem ? problem : null;

  function resetFeedback(): void {
    setError(null);
    setSuccess(null);
  }

  function changeSide(next: OrderSide): void {
    setSide(next);
    setValue('');
    setTouched(false);
    resetFeedback();
  }

  function changeSellMode(next: SellMode): void {
    setSellMode(next);
    setValue('');
    setTouched(false);
    resetFeedback();
  }

  async function handleReview(): Promise<void> {
    resetFeedback();
    setTouched(true);
    if (problem) return;
    if (riskMismatch && !riskAck) {
      setError(t('errors.RISK_ACK_REQUIRED'));
      return;
    }

    if (isGold) {
      setSubmitting(true);
      setQuoteError(null);
      try {
        setQuote(await orderService.requestGoldQuote(goldQuoteBody()));
      } catch (err) {
        setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
      } finally {
        setSubmitting(false);
      }
      return;
    }

    // Fund / Bond direct order. Sell can be by units or by THB.
    setSubmitting(true);
    try {
      const order = await orderService.placeOrder({
        assetType: asset.assetType,
        assetId: asset.id,
        side,
        orderMode: byQty ? 'units' : 'amount',
        ...(byQty ? { units: value } : { amount: value }),
        riskAck: riskMismatch ? riskAck : undefined,
        idempotencyKey: `idem-${asset.assetType}-${asset.id}-${Date.now()}`,
      });
      setSuccess(`${t('page.detail.orderPlaced')} · ${order.txnRef}`);
      setValue('');
      setTouched(false);
      await refreshAfterOrder();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  function goldQuoteBody(): Parameters<typeof orderService.requestGoldQuote>[0] {
    return {
      assetType: 'gold',
      assetId: asset.id,
      side,
      ...(byQty ? { weightGrams: value } : { amount: value }),
    };
  }

  async function refreshAfterOrder(): Promise<void> {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['wallet'] }),
      queryClient.invalidateQueries({ queryKey: ['portfolio'] }),
    ]);
  }

  async function handleConfirmQuote(): Promise<void> {
    if (!quote) return;
    setSubmitting(true);
    setQuoteError(null);
    try {
      const order = await orderService.confirmGoldQuote(
        quote.quoteId,
        `idem-gold-${asset.id}-${quote.quoteId}`,
      );
      setQuote(null);
      setSuccess(`${t('page.detail.orderPlaced')} · ${order.txnRef}`);
      setValue('');
      setTouched(false);
      await refreshAfterOrder();
    } catch (err) {
      setQuoteError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRequote(): Promise<void> {
    setQuoteError(null);
    try {
      setQuote(await orderService.requestGoldQuote(goldQuoteBody()));
    } catch (err) {
      setQuoteError(err instanceof ApiClientError ? err.message : t('errors.UNKNOWN'));
    }
  }

  function goToDca(): void {
    navigate('/dca', {
      state: { assetType: asset.assetType, assetId: asset.id, assetCode: asset.code },
    });
  }

  const tradeDisabled = !priceAvailable || submitting;
  const hint =
    side === 'buy'
      ? isFund && minBuy > 0
        ? t('trade.hintMin').replace('{min}', formatTHB(minBuy))
        : isBond && lot > 0
          ? t('trade.hintLot').replace('{lot}', formatTHB(lot))
          : undefined
      : undefined;

  return (
    <div className="flex flex-col gap-4">
      {/* Buy / Sell toggle */}
      <div
        role="group"
        aria-label="Order side"
        className="grid grid-cols-2 gap-1 rounded-full bg-surface-muted p-1"
      >
        {(['buy', 'sell'] as const).map((s) => (
          <button
            key={s}
            type="button"
            onClick={() => changeSide(s)}
            aria-pressed={side === s}
            className={[
              'min-h-[44px] rounded-full text-sm font-semibold transition-standard duration-standard',
              side === s
                ? 'bg-primary text-text-inverse shadow-sm'
                : 'text-text-muted hover:text-primary',
            ].join(' ')}
          >
            {s === 'buy' ? t('page.detail.buy') : t('page.detail.sell')}
          </button>
        ))}
      </div>

      {!priceAvailable && (
        <p
          role="alert"
          className="rounded-md border border-danger bg-danger/10 p-3 text-sm text-danger"
        >
          {t('page.detail.priceUnavailable')}
        </p>
      )}

      {side === 'sell' && (
        <div className="rounded-md bg-surface-muted px-4 py-3 text-sm">
          <span className="text-text-muted">{t('trade.youHold')} </span>
          <span className="font-semibold text-text">
            {heldQty > 0 ? `${formatUnits(heldQty)} ${qtyUnit} ≈ ${formatTHB(heldValue)}` : '—'}
          </span>
        </div>
      )}

      {side === 'sell' && heldQty > 0 && (
        <div role="group" aria-label="Sell by" className="flex gap-2">
          {(['amount', 'qty'] as const).map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => changeSellMode(m)}
              aria-pressed={sellMode === m}
              className={[
                'min-h-[44px] flex-1 rounded-full border text-sm font-medium transition-standard duration-standard',
                sellMode === m
                  ? 'border-primary bg-primary-soft text-primary'
                  : 'border-border text-text-muted hover:border-primary hover:text-primary',
              ].join(' ')}
            >
              {m === 'amount'
                ? t('trade.byAmount')
                : isGold
                  ? t('trade.byGrams')
                  : t('trade.byUnits')}
            </button>
          ))}
        </div>
      )}

      <AmountInput
        label={side === 'buy' ? t('trade.buyAmountLabel') : t('trade.sellAmountLabel')}
        unit={byQty ? qtyUnit : t('trade.baht')}
        value={value}
        decimals={byQty ? 4 : 2}
        onChange={(v) => {
          setValue(v);
          setTouched(v !== '');
          resetFeedback();
        }}
        picks={picks}
        picksLabel={t('trade.quickPicks')}
        hint={hint}
        error={showProblem}
        disabled={!priceAvailable || (side === 'sell' && heldQty <= 0)}
      />

      <dl className="space-y-1 rounded-md border border-border px-4 py-3 text-sm text-text-muted">
        {byQty ? (
          <div className="flex justify-between">
            <dt>{t('trade.estProceeds')}</dt>
            <dd className="font-semibold text-text">{formatTHB(orderThb)}</dd>
          </div>
        ) : (
          !isBond && (
            <div className="flex justify-between">
              <dt>{isGold ? t('trade.estGrams') : t('page.detail.estUnits')}</dt>
              <dd className="font-semibold text-text">
                {isGold
                  ? `${formatUnits(orderQty)} ${t('trade.gramsSuffix')}`
                  : formatUnits(orderQty)}
              </dd>
            </div>
          )
        )}
        <div className="flex justify-between">
          <dt>{t('page.detail.cashAvailable')}</dt>
          <dd className="text-text">{formatTHB(available)}</dd>
        </div>
      </dl>

      {riskMismatch && (
        <RiskAckBox
          assetRisk={asset.riskLevel}
          userRisk={userRisk}
          checked={riskAck}
          onChange={setRiskAck}
        />
      )}

      {error && (
        <p role="alert" className="text-sm text-danger">
          {error}
        </p>
      )}
      {success && (
        <p
          role="status"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-success"
        >
          <CheckCircle size={16} />
          {success}
        </p>
      )}

      <Button variant="primary" onClick={() => void handleReview()} disabled={tradeDisabled}>
        {submitting
          ? t('common.loading')
          : isGold
            ? t('page.detail.getQuote')
            : t('page.detail.reviewOrder')}
      </Button>

      {!isBond && (
        <Button variant="outline" onClick={goToDca}>
          {t('page.detail.setupDca')}
        </Button>
      )}

      {quote && (
        <QuoteDialog
          quote={quote}
          side={side}
          submitting={submitting}
          error={quoteError}
          onConfirm={handleConfirmQuote}
          onRequote={handleRequote}
          onClose={() => {
            setQuote(null);
            setQuoteError(null);
          }}
        />
      )}
    </div>
  );
}
