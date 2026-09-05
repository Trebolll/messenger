import { type ReactNode } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { useAuth } from '../features/auth/AuthContext';
import { LoginPage } from '../features/auth/LoginPage';
import { RegisterPage } from '../features/auth/RegisterPage';
import { ChatPage } from '../features/chats/ChatPage';
import { RealtimeProvider } from '../features/realtime/RealtimeContext';

function Protected({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  return <RealtimeProvider enabled>{children}</RealtimeProvider>;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/" element={<Protected><ChatPage /></Protected>} />
    </Routes>
  );
}
