import { Navigate, Outlet } from 'react-router-dom';
import { isAuthenticated } from '../features/auth/api';

export default function AuthGuard() {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}
