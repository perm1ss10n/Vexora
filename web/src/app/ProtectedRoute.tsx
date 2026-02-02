import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth';

export function ProtectedRoute() {
  const { accessToken, status } = useAuthStore((state) => ({
    accessToken: state.accessToken,
    status: state.status,
  }));
  const location = useLocation();

  if (status === 'loading' || status === 'idle') {
    return (
      <div className="grid min-h-screen place-items-center bg-background text-sm text-muted-foreground">
        Loading secure session...
      </div>
    );
  }

  if (!accessToken) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <Outlet />;
}
