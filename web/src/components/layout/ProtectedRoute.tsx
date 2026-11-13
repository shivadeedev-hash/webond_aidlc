// Route guard: redirects guests to Login, preserving the attempted path so the
// user returns there after signing in (US-005).

import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { LoadingState } from '../ui/StateViews';

export default function ProtectedRoute(): JSX.Element {
  const { isAuthenticated, isInitializing } = useAuth();
  const location = useLocation();

  // A saved session may still be restoring after a reload: wait instead of
  // bouncing a signed-in user to the login page.
  if (isInitializing) {
    return <LoadingState />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <Outlet />;
}
