import { useEffect } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { PublicLayout } from './layouts/PublicLayout';
import { AppLayout } from './layouts/AppLayout';
import { ProtectedRoute } from './ProtectedRoute';
import { HomePage } from '@/pages/HomePage';
import { AboutPage } from '@/pages/AboutPage';
import { SpecPage } from '@/pages/SpecPage';
import { OrderPage } from '@/pages/OrderPage';
import { SupportPage } from '@/pages/SupportPage';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { DevicesPage } from '@/pages/DevicesPage';
import { DeviceDetailPage } from '@/pages/DeviceDetailPage';
import { TelemetryPage } from '@/pages/TelemetryPage';
import { CommandsPage } from '@/pages/CommandsPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { useAuthStore } from '@/store/auth';

export default function App() {
  const bootstrap = useAuthStore((state) => state.bootstrap);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route index element={<HomePage />} />
        <Route path="/about" element={<AboutPage />} />
        <Route path="/spec" element={<SpecPage />} />
        <Route path="/order" element={<OrderPage />} />
        <Route path="/support" element={<SupportPage />} />
      </Route>

      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      <Route path="/app" element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/app/devices" replace />} />
          <Route path="devices" element={<DevicesPage />} />
          <Route path="devices/:deviceId" element={<DeviceDetailPage />} />
          <Route path="telemetry" element={<TelemetryPage />} />
          <Route path="commands" element={<CommandsPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
      </Route>
    </Routes>
  );
}
