// App shell: sticky white header + a centered main content region that hosts
// the routed page via <Outlet>.

import { Outlet } from 'react-router-dom';
import Header from './Header';
import Container from './Container';

export default function AppLayout(): JSX.Element {
  return (
    <div className="flex min-h-full flex-col bg-bg text-text">
      <Header />
      <main className="flex-1 py-section">
        <Container>
          <Outlet />
        </Container>
      </main>
    </div>
  );
}
