// DCA Plans (US-019..021, wireframe 05). Plan table + create/edit form (MF & Gold
// only, bond excluded BR-06) + round history per plan. Min amount 500 MF / 100
// Gold enforced in the form (BR-07) and re-checked by the service.

import { useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { dcaService } from '../services/dca';
import { catalogService } from '../services/catalog';
import type { CreateDcaPlanRequest, DcaAssetType, DcaPlan, UpdateDcaPlanRequest } from '../types';
import { useI18n } from '../i18n/useI18n';
import { formatDateBangkok } from '../lib/format';
import PageHeader from '../components/ui/PageHeader';
import Button from '../components/ui/Button';
import { Card, CardBody, CardHeader, CardTitle } from '../components/ui/Card';
import { LoadingState, ErrorState, EmptyState } from '../components/ui/StateViews';
import StatusBadge from '../components/ui/StatusBadge';
import PlanTable from '../components/dca/PlanTable';
import PlanForm from '../components/dca/PlanForm';

interface DcaNavState {
  assetType?: DcaAssetType;
  assetId?: number;
}

export default function DCA(): JSX.Element {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const location = useLocation();
  const navState = (location.state as DcaNavState | null) ?? null;

  const [showForm, setShowForm] = useState<boolean>(navState?.assetId !== undefined);
  const [editing, setEditing] = useState<DcaPlan | null>(null);
  const [expandedPlan, setExpandedPlan] = useState<number | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);

  const plansQuery = useQuery({ queryKey: ['dcaPlans'], queryFn: () => dcaService.listPlans() });

  // Fund + gold assets for the picker and for resolving codes in the table.
  const fundsQuery = useQuery({
    queryKey: ['market', { type: 'fund' }],
    queryFn: () => catalogService.listAssets({ type: 'fund' }),
  });
  const goldQuery = useQuery({
    queryKey: ['market', { type: 'gold' }],
    queryFn: () => catalogService.listAssets({ type: 'gold' }),
  });

  const dcaAssets = useMemo(
    () => [...(fundsQuery.data?.items ?? []), ...(goldQuery.data?.items ?? [])],
    [fundsQuery.data, goldQuery.data],
  );

  const codeFor = (assetType: string, assetId: number): string =>
    dcaAssets.find((a) => a.assetType === assetType && a.id === assetId)?.code ?? `#${assetId}`;

  const roundsQuery = useQuery({
    queryKey: ['dcaRounds', expandedPlan],
    queryFn: () => dcaService.listRounds(expandedPlan!),
    enabled: expandedPlan !== null,
  });

  async function refresh(): Promise<void> {
    await queryClient.invalidateQueries({ queryKey: ['dcaPlans'] });
  }

  async function handleCreate(req: CreateDcaPlanRequest): Promise<void> {
    await dcaService.createPlan(req);
    await refresh();
    setShowForm(false);
  }

  async function handleUpdate(id: number, req: UpdateDcaPlanRequest): Promise<void> {
    await dcaService.updatePlan(id, req);
    await refresh();
    setEditing(null);
    setShowForm(false);
  }

  async function runAction(id: number, fn: (id: number) => Promise<unknown>): Promise<void> {
    setBusyId(id);
    try {
      await fn(id);
      await refresh();
    } finally {
      setBusyId(null);
    }
  }

  if (plansQuery.isLoading) return <LoadingState />;
  if (plansQuery.isError) return <ErrorState onRetry={() => void plansQuery.refetch()} />;

  const plans = plansQuery.data ?? [];
  const preset =
    navState?.assetType && navState.assetId
      ? { assetType: navState.assetType, assetId: navState.assetId }
      : null;

  return (
    <div>
      <PageHeader
        title={t('page.dca.title')}
        action={
          !showForm && (
            <Button
              variant="primary"
              onClick={() => {
                setEditing(null);
                setShowForm(true);
              }}
            >
              {t('page.dca.new')}
            </Button>
          )
        }
      />

      <Card className="mb-6">
        <CardBody className="p-0">
          {plans.length === 0 ? (
            <EmptyState />
          ) : (
            <PlanTable
              plans={plans}
              codeFor={codeFor}
              busyId={busyId}
              onEdit={(plan) => {
                setEditing(plan);
                setShowForm(true);
              }}
              onPause={(id) => void runAction(id, dcaService.pausePlan.bind(dcaService))}
              onResume={(id) => void runAction(id, dcaService.resumePlan.bind(dcaService))}
              onCancel={(id) => void runAction(id, dcaService.cancelPlan.bind(dcaService))}
            />
          )}
        </CardBody>
      </Card>

      {showForm && (
        <Card className="mb-6">
          <CardBody>
            <PlanForm
              assets={dcaAssets}
              editing={editing}
              preset={preset}
              onCreate={handleCreate}
              onUpdate={handleUpdate}
              onCancel={() => {
                setShowForm(false);
                setEditing(null);
              }}
            />
          </CardBody>
        </Card>
      )}

      {/* Round history */}
      <Card>
        <CardHeader>
          <CardTitle>{t('page.dca.rounds')}</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="mb-3 flex flex-wrap gap-2">
            {plans.map((p) => (
              <button
                key={p.id}
                type="button"
                aria-pressed={expandedPlan === p.id}
                onClick={() => setExpandedPlan(expandedPlan === p.id ? null : p.id)}
                className={[
                  'min-h-[40px] rounded-md border px-3 text-sm font-medium transition-standard duration-standard',
                  expandedPlan === p.id
                    ? 'border-primary bg-primary text-text-inverse'
                    : 'border-border hover:bg-surface-muted',
                ].join(' ')}
              >
                {codeFor(p.assetType, p.assetId)}
              </button>
            ))}
          </div>

          {expandedPlan === null ? (
            <p className="text-sm text-text-muted">Select a plan to view its round history.</p>
          ) : roundsQuery.isLoading ? (
            <LoadingState />
          ) : (roundsQuery.data ?? []).length === 0 ? (
            <EmptyState />
          ) : (
            <ul className="flex flex-col divide-y divide-border text-sm">
              {(roundsQuery.data ?? []).map((r) => (
                <li key={r.id} className="flex items-center justify-between gap-3 py-2">
                  <span className="text-text-muted">{formatDateBangkok(r.runDate)}</span>
                  <StatusBadge status={r.status === 'success' ? 'completed' : 'failed'} />
                  {r.failureReason && (
                    <span className="text-xs text-danger">{r.failureReason}</span>
                  )}
                </li>
              ))}
            </ul>
          )}
        </CardBody>
      </Card>
    </div>
  );
}
