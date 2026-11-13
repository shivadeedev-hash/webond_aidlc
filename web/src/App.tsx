// Route map (components.md AppShell). Public: Market (+ asset detail), Login.
// Everything else is guarded by ProtectedRoute (guests → Login, US-005).
// `/` redirects to Market so guests land on a public page.

import { Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './components/layout/AppLayout';
import ProtectedRoute from './components/layout/ProtectedRoute';
import Login from './pages/Login';
import Register from './pages/Register';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import VerifyEmail from './pages/VerifyEmail';
import Market from './pages/Market';
import AssetDetail from './pages/AssetDetail';
import Dashboard from './pages/Dashboard';
import Portfolio from './pages/Portfolio';
import DCA from './pages/DCA';
import History from './pages/History';
import TransactionDetail from './pages/TransactionDetail';
import Profile from './pages/Profile';
import PagePlaceholder from './pages/PagePlaceholder';

export default function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        {/* Public */}
        <Route index element={<Navigate to="/market" replace />} />
        <Route path="market" element={<Market />} />
        <Route path="assets/:type/:id" element={<AssetDetail />} />
        <Route path="login" element={<Login />} />
        <Route path="register" element={<Register />} />
        <Route path="forgot-password" element={<ForgotPassword />} />
        <Route path="reset-password" element={<ResetPassword />} />
        <Route path="verify-email" element={<VerifyEmail />} />

        {/* Protected (US-005) */}
        <Route element={<ProtectedRoute />}>
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="portfolio" element={<Portfolio />} />
          <Route path="dca" element={<DCA />} />
          <Route path="history" element={<History />} />
          <Route path="history/:txnRef" element={<TransactionDetail />} />
          <Route path="profile" element={<Profile />} />
        </Route>

        {/* Fallback */}
        <Route path="*" element={<PagePlaceholder title="404" description="Page not found" />} />
      </Route>
    </Routes>
  );
}
