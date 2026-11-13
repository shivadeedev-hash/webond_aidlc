// Helpers shared by the mock service implementations.
// Mock responses are deep-cloned so callers can't mutate the seed data, and a
// tiny artificial latency keeps loading states realistic in the UI.

export const MOCK_LATENCY_MS = 120;

export function delay(ms: number = MOCK_LATENCY_MS): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** Resolve with a deep clone of the value after a short latency. */
export async function mockResolve<T>(value: T, ms: number = MOCK_LATENCY_MS): Promise<T> {
  await delay(ms);
  return structuredClone(value);
}

/** Simple paginator over an in-memory array. */
export function paginate<T>(
  items: T[],
  page = 1,
  pageSize = 20,
): { items: T[]; meta: { page: number; pageSize: number; total: number; hasMore: boolean } } {
  const start = (page - 1) * pageSize;
  const slice = items.slice(start, start + pageSize);
  return {
    items: slice,
    meta: { page, pageSize, total: items.length, hasMore: start + pageSize < items.length },
  };
}
