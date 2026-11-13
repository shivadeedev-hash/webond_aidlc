// Transaction/order status badge. Pairs a semantic color with an icon AND a text
// label so meaning never relies on color alone (design-system §2.1, WCAG 2.2 AA).

import type { TransactionStatus } from '../../types';
import { CheckCircle, Clock, XCircle, Ban } from './icons';

type AnyStatus = TransactionStatus | 'pending';

interface Spec {
  label: string;
  className: string;
  Icon: (p: { size?: number }) => JSX.Element;
}

const SPECS: Record<AnyStatus, Spec> = {
  submitted: { label: 'Submitted', className: 'text-info', Icon: Clock },
  processing: { label: 'Processing', className: 'text-warning', Icon: Clock },
  pending: { label: 'Pending', className: 'text-warning', Icon: Clock },
  completed: { label: 'Completed', className: 'text-success', Icon: CheckCircle },
  failed: { label: 'Failed', className: 'text-danger', Icon: XCircle },
  cancelled: { label: 'Cancelled', className: 'text-text-muted', Icon: Ban },
};

export default function StatusBadge({
  status,
  suffix,
}: {
  status: AnyStatus;
  suffix?: string;
}): JSX.Element {
  const spec = SPECS[status] ?? SPECS.submitted;
  const { Icon } = spec;
  return (
    <span className={`inline-flex items-center gap-1.5 text-sm font-medium ${spec.className}`}>
      <Icon size={16} />
      <span>
        {spec.label}
        {suffix ? ` ${suffix}` : ''}
      </span>
    </span>
  );
}
